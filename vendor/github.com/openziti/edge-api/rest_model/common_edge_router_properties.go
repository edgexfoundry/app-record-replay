// Code generated by go-swagger; DO NOT EDIT.

//
// Copyright NetFoundry Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// __          __              _
// \ \        / /             (_)
//  \ \  /\  / /_ _ _ __ _ __  _ _ __   __ _
//   \ \/  \/ / _` | '__| '_ \| | '_ \ / _` |
//    \  /\  / (_| | |  | | | | | | | | (_| | : This file is generated, do not edit it.
//     \/  \/ \__,_|_|  |_| |_|_|_| |_|\__, |
//                                      __/ |
//                                     |___/

package rest_model

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
	"github.com/go-openapi/validate"
)

// CommonEdgeRouterProperties common edge router properties
//
// swagger:model commonEdgeRouterProperties
type CommonEdgeRouterProperties struct {

	// app data
	AppData *Tags `json:"appData,omitempty"`

	// cost
	// Required: true
	// Maximum: 65535
	// Minimum: 0
	Cost *int64 `json:"cost"`

	// disabled
	// Required: true
	Disabled *bool `json:"disabled"`

	// hostname
	// Required: true
	Hostname *string `json:"hostname"`

	// is online
	// Required: true
	IsOnline *bool `json:"isOnline"`

	// name
	// Required: true
	Name *string `json:"name"`

	// no traversal
	// Required: true
	NoTraversal *bool `json:"noTraversal"`

	// supported protocols
	// Required: true
	SupportedProtocols map[string]string `json:"supportedProtocols"`

	// sync status
	// Required: true
	SyncStatus *string `json:"syncStatus"`
}

// Validate validates this common edge router properties
func (m *CommonEdgeRouterProperties) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateAppData(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateCost(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateDisabled(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateHostname(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateIsOnline(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateName(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateNoTraversal(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateSupportedProtocols(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateSyncStatus(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *CommonEdgeRouterProperties) validateAppData(formats strfmt.Registry) error {
	if swag.IsZero(m.AppData) { // not required
		return nil
	}

	if m.AppData != nil {
		if err := m.AppData.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("appData")
			} else if ce, ok := err.(*errors.CompositeError); ok {
				return ce.ValidateName("appData")
			}
			return err
		}
	}

	return nil
}

func (m *CommonEdgeRouterProperties) validateCost(formats strfmt.Registry) error {

	if err := validate.Required("cost", "body", m.Cost); err != nil {
		return err
	}

	if err := validate.MinimumInt("cost", "body", *m.Cost, 0, false); err != nil {
		return err
	}

	if err := validate.MaximumInt("cost", "body", *m.Cost, 65535, false); err != nil {
		return err
	}

	return nil
}

func (m *CommonEdgeRouterProperties) validateDisabled(formats strfmt.Registry) error {

	if err := validate.Required("disabled", "body", m.Disabled); err != nil {
		return err
	}

	return nil
}

func (m *CommonEdgeRouterProperties) validateHostname(formats strfmt.Registry) error {

	if err := validate.Required("hostname", "body", m.Hostname); err != nil {
		return err
	}

	return nil
}

func (m *CommonEdgeRouterProperties) validateIsOnline(formats strfmt.Registry) error {

	if err := validate.Required("isOnline", "body", m.IsOnline); err != nil {
		return err
	}

	return nil
}

func (m *CommonEdgeRouterProperties) validateName(formats strfmt.Registry) error {

	if err := validate.Required("name", "body", m.Name); err != nil {
		return err
	}

	return nil
}

func (m *CommonEdgeRouterProperties) validateNoTraversal(formats strfmt.Registry) error {

	if err := validate.Required("noTraversal", "body", m.NoTraversal); err != nil {
		return err
	}

	return nil
}

func (m *CommonEdgeRouterProperties) validateSupportedProtocols(formats strfmt.Registry) error {

	if err := validate.Required("supportedProtocols", "body", m.SupportedProtocols); err != nil {
		return err
	}

	return nil
}

func (m *CommonEdgeRouterProperties) validateSyncStatus(formats strfmt.Registry) error {

	if err := validate.Required("syncStatus", "body", m.SyncStatus); err != nil {
		return err
	}

	return nil
}

// ContextValidate validate this common edge router properties based on the context it is used
func (m *CommonEdgeRouterProperties) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	var res []error

	if err := m.contextValidateAppData(ctx, formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *CommonEdgeRouterProperties) contextValidateAppData(ctx context.Context, formats strfmt.Registry) error {

	if m.AppData != nil {
		if err := m.AppData.ContextValidate(ctx, formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("appData")
			} else if ce, ok := err.(*errors.CompositeError); ok {
				return ce.ValidateName("appData")
			}
			return err
		}
	}

	return nil
}

// MarshalBinary interface implementation
func (m *CommonEdgeRouterProperties) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *CommonEdgeRouterProperties) UnmarshalBinary(b []byte) error {
	var res CommonEdgeRouterProperties
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
