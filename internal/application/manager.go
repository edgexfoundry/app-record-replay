//
// Copyright (c) 2023 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package application

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	appInterfaces "github.com/edgexfoundry/app-functions-sdk-go/v3/pkg/interfaces"
	"github.com/edgexfoundry/app-functions-sdk-go/v3/pkg/transforms"
	"github.com/edgexfoundry/app-record-replay/internal/interfaces"
	"github.com/edgexfoundry/app-record-replay/pkg/dtos"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/utils"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/common"
	coreDtos "github.com/edgexfoundry/go-mod-core-contracts/v3/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/dtos/requests"
	"github.com/google/uuid"
)

const (
	createBatchFailedMessage           = "failed to create Batch pipeline function"
	setPipelineFailedMessage           = "failed to set the default function pipeline"
	debugFilterMessage                 = "ARR Start Recording: Filter %s names %v function added to the functions pipeline"
	debugPipelineFunctionsAddedMessage = "ARR Start Recording: CountEvents, Batch and ProcessBatchedData functions added to the functions pipeline"
	replayExiting                      = "ARR Replay: Replay exiting due to App termination"
	replayPublishFailed                = "failed to publish replay event: %v"
	replayDeepCopyFailed               = "deep copy of event to be replayed failed: %v"
	maxReplayDelayExceeded             = "%s delay exceeds the maximum replay delay of %s. Maximum replay delay is configurable using MaxReplayDelay App Setting"
	noReplayExists                     = "no replay running or has previously been run"
)

var recordingInProgressError = errors.New("a recording is in progress")
var batchParametersNotSetError = errors.New("duration and/or count not set")
var noRecordingRunningToCancelError = errors.New("no recording currently running")

var replayInProgressError = errors.New("a replay is in progress")
var noRecordedData = errors.New("no recorded data present")
var invalidReplayRate = errors.New("invalid ReplayRate, value must be greater than 0")
var invalidReplayCount = errors.New("invalid ReplayCount, value must be greater than or equal 0. Zero defaults to 1")

var replayCanceled = errors.New("replay canceled")

type recordedData struct {
	Duration time.Duration
	Events   []coreDtos.Event
	Devices  []coreDtos.Device
	Profiles []coreDtos.DeviceProfile
}

// dataManager implements interface that records and replays captured data
type dataManager struct {
	appSvc         appInterfaces.ApplicationService
	recordingMutex sync.Mutex

	recordedEventCount int
	recordingStartedAt *time.Time
	recordedData       *recordedData

	maxReplayDelay      time.Duration
	replayStartedAt     *time.Time
	replayedDuration    time.Duration
	replayedEventCount  int
	replayedRepeatCount int
	replayError         error
	replayContext       context.Context
	replayCancelFunc    context.CancelFunc
}

// NewManager is the factory function which instantiates a Data Manager
func NewManager(service appInterfaces.ApplicationService, maxReplayDelay time.Duration) interfaces.DataManager {
	return &dataManager{
		appSvc:         service,
		maxReplayDelay: maxReplayDelay,
	}
}

// StartRecording starts a recording session based on the values in the request.
// An error is returned if the request data is incomplete or a record or replay session is currently running.
func (m *dataManager) StartRecording(request dtos.RecordRequest) error {
	lc := m.appSvc.LoggingClient()

	m.recordingMutex.Lock()
	defer m.recordingMutex.Unlock()

	if m.recordingStartedAt != nil {
		return recordingInProgressError
	}

	if m.replayStartedAt != nil {
		return replayInProgressError
	}

	m.recordedData = nil
	m.recordedEventCount = 0

	var pipeline []appInterfaces.AppFunction

	if len(request.IncludeDeviceProfiles) > 0 {
		includeFilter := transforms.NewFilterFor(request.IncludeDeviceProfiles)
		pipeline = append(pipeline, includeFilter.FilterByProfileName)
		lc.Debugf(debugFilterMessage, "for profile", request.IncludeDeviceProfiles)
	}

	if len(request.ExcludeDeviceProfiles) > 0 {
		excludeFilter := transforms.NewFilterOut(request.ExcludeDeviceProfiles)
		pipeline = append(pipeline, excludeFilter.FilterByProfileName)
		lc.Debugf(debugFilterMessage, "out profile", request.ExcludeDeviceProfiles)
	}

	if len(request.IncludeDevices) > 0 {
		includeFilter := transforms.NewFilterFor(request.IncludeDevices)
		pipeline = append(pipeline, includeFilter.FilterByDeviceName)
		lc.Debugf(debugFilterMessage, "for device", request.IncludeDevices)
	}

	if len(request.ExcludeDevices) > 0 {
		excludeFilter := transforms.NewFilterOut(request.ExcludeDevices)
		pipeline = append(pipeline, excludeFilter.FilterByDeviceName)
		lc.Debugf(debugFilterMessage, "out device", request.ExcludeDevices)
	}

	if len(request.IncludeSources) > 0 {
		includeFilter := transforms.NewFilterFor(request.IncludeSources)
		pipeline = append(pipeline, includeFilter.FilterBySourceName)
		lc.Debugf(debugFilterMessage, "for source", request.IncludeSources)
	}

	if len(request.ExcludeSources) > 0 {
		excludeFilter := transforms.NewFilterOut(request.ExcludeSources)
		pipeline = append(pipeline, excludeFilter.FilterBySourceName)
		lc.Debugf(debugFilterMessage, "out source", request.ExcludeSources)
	}

	var batch *transforms.BatchConfig
	var err error

	if request.Duration > 0 && request.EventLimit > 0 {
		batch, err = transforms.NewBatchByTimeAndCount(request.Duration.String(), request.EventLimit)
	} else if request.EventLimit > 0 {
		batch, err = transforms.NewBatchByCount(request.EventLimit)
	} else if request.Duration > 0 {
		batch, err = transforms.NewBatchByTime(request.Duration.String())
	} else {
		err = batchParametersNotSetError
	}

	if err != nil {
		return fmt.Errorf("%s: %v", createBatchFailedMessage, err)
	}

	// processBatchedData expects slice of Events, so configure batch to return slice of Events
	batch.IsEventData = true

	pipeline = append(pipeline, m.countEvents, batch.Batch, m.processBatchedData)
	lc.Debug(debugPipelineFunctionsAddedMessage)

	// Setting the Functions Pipeline starts the recording of Events
	err = m.appSvc.SetDefaultFunctionsPipeline(pipeline...)
	if err != nil {
		return fmt.Errorf("%s: %v", setPipelineFailedMessage, err)
	}

	now := time.Now()
	m.recordingStartedAt = &now

	lc.Debugf("ARR Start Recording: Recording of Events has started with EventLimit=%d and Duration=%s", request.EventLimit, request.Duration.String())

	return nil
}

// CancelRecording cancels the current recording session
func (m *dataManager) CancelRecording() error {
	m.recordingMutex.Lock()
	defer m.recordingMutex.Unlock()

	if m.recordingStartedAt == nil {
		return noRecordingRunningToCancelError
	}

	// This stops recording of Events
	m.appSvc.RemoveAllFunctionPipelines()
	m.recordingStartedAt = nil

	m.appSvc.LoggingClient().Debug("ARR Cancel Recording: Recording of Events has been canceled")

	return nil
}

// RecordingStatus returns the status of the current recording session
func (m *dataManager) RecordingStatus() dtos.RecordStatus {
	m.recordingMutex.Lock()
	defer m.recordingMutex.Unlock()

	status := dtos.RecordStatus{}

	if m.recordingStartedAt != nil {
		status.InProgress = true
		status.Duration = time.Since(*m.recordingStartedAt)
		status.EventCount = m.recordedEventCount
	} else if m.recordedData != nil {
		status.Duration = m.recordedData.Duration
		status.EventCount = len(m.recordedData.Events)
	}

	return status
}

// StartReplay starts a replay session based on the values in the request
// An error is returned if the request data is incomplete or a record or replay session is currently running.
func (m *dataManager) StartReplay(request dtos.ReplayRequest) error {
	m.recordingMutex.Lock()
	defer m.recordingMutex.Unlock()

	if m.recordingStartedAt != nil {
		return recordingInProgressError
	}

	if m.replayStartedAt != nil {
		return replayInProgressError
	}

	if m.recordedData == nil {
		return noRecordedData
	}

	if request.ReplayRate <= 0 {
		return invalidReplayRate
	}

	if request.RepeatCount < 0 {
		return invalidReplayCount
	}

	now := time.Now()
	m.replayStartedAt = &now
	m.replayedDuration = 0
	m.replayedEventCount = 0
	m.replayedRepeatCount = 0
	m.replayError = nil
	m.replayContext, m.replayCancelFunc = context.WithCancel(context.Background())

	go m.replayRecordedEvents(request)

	return nil
}

func (m *dataManager) replayRecordedEvents(request dtos.ReplayRequest) {
	var previousEventTime int64
	firstEvent := true
	lc := m.appSvc.LoggingClient()

	// Replay Count of zero defaults to 1.
	replayCount := 1
	if request.RepeatCount > 0 {
		replayCount = request.RepeatCount
	}

	lc.Debugf("ARR Replay: Replay starting with Replay Rate of %v and Repeat Count of %d ", request.ReplayRate, replayCount)

	for i := 0; i < replayCount; i++ {
		for _, event := range m.recordedData.Events {
			// Check if service is terminating
			if m.appSvc.AppContext().Err() != nil {
				m.recordingMutex.Lock()
				m.replayStartedAt = nil
				m.recordingMutex.Unlock()
				m.appSvc.LoggingClient().Info(replayExiting)
				return
			}

			// Check if replay cancel func has been called to cancel the replay
			if m.replayContext.Err() != nil {
				m.setReplayError(replayCanceled)
				return
			}

			replayEvent := coreDtos.Event{}
			if err := utils.DeepCopy(event, &replayEvent); err != nil {
				m.setReplayError(fmt.Errorf(replayDeepCopyFailed, err))
				return
			}

			// Send the first event immediately and then wait appropriate time between events
			if firstEvent {
				firstEvent = false
			} else {
				delay := replayEvent.Origin - previousEventTime

				// Replay Rate less than one increases the delay to slow down replay pace while greater than one
				// decreases the delay to increase the replay pace.
				delay = int64(float32(delay) * (1 / request.ReplayRate))

				if time.Duration(delay) > m.maxReplayDelay {
					m.setReplayError(fmt.Errorf(maxReplayDelayExceeded, time.Duration(delay).String(), m.maxReplayDelay.String()))
					return
				}

				// Best we can do with realtime capabilities
				time.Sleep(time.Duration(delay))
			}

			previousEventTime = replayEvent.Origin

			topic := common.BuildTopic(strings.Replace(common.CoreDataEventSubscribeTopic, "/#", "", 1),
				replayEvent.ProfileName, replayEvent.DeviceName, replayEvent.SourceName)

			newOrigin := time.Now().UnixNano()
			replayEvent.Origin = newOrigin
			replayEvent.Id = uuid.NewString()
			for index := range replayEvent.Readings {
				replayEvent.Readings[index].Origin = newOrigin
				replayEvent.Readings[index].Id = uuid.NewString()
			}

			addEvent := requests.NewAddEventRequest(replayEvent)

			if err := m.appSvc.PublishWithTopic(topic, addEvent, common.ContentTypeJSON); err != nil {
				m.setReplayError(fmt.Errorf(replayPublishFailed, err))
				return
			}

			lc.Debugf("ARR Replay: Replayed Event to topic: %s", topic)

			m.incrementReplayedEventCount()
		}

		m.incrementReplayRepeatCount()
	}

	m.recordingMutex.Lock()
	defer m.recordingMutex.Unlock()
	m.replayedDuration = time.Since(*m.replayStartedAt)
	m.replayStartedAt = nil

	lc.Debugf("ARR Replay: Replay completed in %s. %d events replayed with %d repeated replays",
		m.replayedDuration.String(), m.replayedEventCount, m.replayedRepeatCount)
}

func (m *dataManager) setReplayError(err error) {
	m.recordingMutex.Lock()
	defer m.recordingMutex.Unlock()
	m.replayError = err
	m.replayStartedAt = nil
	m.appSvc.LoggingClient().Errorf("ARR Replay: Replay stopped due to error: %v", err)
}

func (m *dataManager) incrementReplayedEventCount() {
	m.recordingMutex.Lock()
	defer m.recordingMutex.Unlock()
	m.replayedEventCount++
}

func (m *dataManager) incrementReplayRepeatCount() {
	m.recordingMutex.Lock()
	defer m.recordingMutex.Unlock()
	m.replayedRepeatCount++
}

// CancelReplay cancels the current replay session
func (m *dataManager) CancelReplay() error {
	//TODO implement me using TDD
	return errors.New("not implemented")
}

// ReplayStatus returns the status of the current replay session
func (m *dataManager) ReplayStatus() dtos.ReplayStatus {
	m.recordingMutex.Lock()
	defer m.recordingMutex.Unlock()

	duration := m.replayedDuration

	// If replay is in progress we need to calculate the duration so far.
	if m.replayedDuration == 0 && m.replayStartedAt != nil {
		duration = time.Since(*m.replayStartedAt)
	}

	errorMessage := ""
	if m.replayError != nil {
		errorMessage = m.replayError.Error()
	} else if m.replayStartedAt == nil && m.replayedDuration == 0 {
		errorMessage = noReplayExists
	}

	return dtos.ReplayStatus{
		Running:      m.replayStartedAt != nil,
		EventCount:   m.replayedEventCount,
		Duration:     duration,
		RepeatCount:  m.replayedRepeatCount,
		ErrorMessage: errorMessage,
	}
}

// ExportRecordedData returns the data for the last record session
// An error is returned if the no record session was run or a record session is currently running
func (m *dataManager) ExportRecordedData() (*dtos.RecordedData, error) {
	//TODO implement me using TDD
	return nil, errors.New("not implemented")
}

// ImportRecordedData imports data from a previously exported record session.
// An error is returned if a record or replay session is currently running or the data is incomplete
func (m *dataManager) ImportRecordedData(data *dtos.RecordedData) error {
	//TODO implement me using TDD
	return errors.New("not implemented")
}

// Pipeline functions

var countsNoDataError = errors.New("CountEvents function received nil data")
var countsDataNotEventError = errors.New("CountEvents function received data that is not an Event")

// countEvents counts the number of Events recorded so far. Must be called after any filters and before the Batch function.
// This count is used when reporting Recording Status
func (m *dataManager) countEvents(_ appInterfaces.AppFunctionContext, data any) (bool, interface{}) {
	if data == nil {
		return false, countsNoDataError
	}

	if _, ok := data.(coreDtos.Event); !ok {
		return false, countsDataNotEventError
	}

	m.recordingMutex.Lock()
	defer m.recordingMutex.Unlock()

	m.recordedEventCount++

	m.appSvc.LoggingClient().Debugf("ARR Event Count: received event to be recorded. Current event count is %d", m.recordedEventCount)

	return true, data
}

var batchNoDataError = errors.New("ProcessBatchedData function received nil data")
var batchDataNotEventCollectionError = errors.New("ProcessBatchedData function received data that is not collection of Event")

// processBatchedData processes the batched data for the current recording session
func (m *dataManager) processBatchedData(_ appInterfaces.AppFunctionContext, data any) (bool, interface{}) {
	lc := m.appSvc.LoggingClient()

	m.recordingMutex.Lock()
	defer m.recordingMutex.Unlock()

	// Check if record was canceled and exit early
	if m.recordingStartedAt == nil {
		return false, nil
	}

	// This stops recording of Events
	m.appSvc.RemoveAllFunctionPipelines()
	lc.Debug("ARR Process Recorded Data: Recording of Events has ended and functions pipeline has been removed")

	if data == nil {
		return false, batchNoDataError
	}

	events, ok := data.([]coreDtos.Event)
	if !ok {
		return false, batchDataNotEventCollectionError
	}

	duration := 0 * time.Second
	if m.recordingStartedAt != nil {
		duration = time.Since(*m.recordingStartedAt)
	}

	m.recordedData = &recordedData{
		Events:   events,
		Duration: duration,
	}

	m.recordingStartedAt = nil

	lc.Debugf("ARR Process Recorded Data: %d events in %s have been saved for replay", len(events), duration.String())

	return false, nil
}
