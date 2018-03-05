// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

// Package javascript contains the Javascript payload formatter message processors.
package javascript

import (
	"context"
	"fmt"
	"reflect"

	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/gogoproto"
	"github.com/TheThingsNetwork/ttn/pkg/messageprocessors"
	"github.com/TheThingsNetwork/ttn/pkg/scripting"
	js "github.com/TheThingsNetwork/ttn/pkg/scripting/javascript"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
)

type host struct {
	engine scripting.Engine
}

// New creates and returns a new Javascript payload encoder and decoder.
func New() messageprocessors.PayloadEncodeDecoder {
	return &host{
		engine: js.New(scripting.DefaultOptions),
	}
}

func (h *host) createEnvironment(model *ttnpb.EndDeviceModel) map[string]interface{} {
	env := make(map[string]interface{})
	env["brand"] = model.Brand
	env["model"] = model.Model
	env["hardware_version"] = model.HardwareVersion
	env["firmware_version"] = model.FirmwareVersion
	return env
}

// Encode encodes the message's MAC payload DecodedPayload to FRMPayload using script.
func (h *host) Encode(ctx context.Context, msg *ttnpb.DownlinkMessage, model *ttnpb.EndDeviceModel, script string) (*ttnpb.DownlinkMessage, error) {
	payload := msg.Payload.GetMACPayload()
	if payload == nil {
		return nil, messageprocessors.ErrNoMACPayload.New(nil)
	}

	decoded := payload.DecodedPayload
	if decoded == nil {
		return msg, nil
	}

	m, err := gogoproto.Map(decoded)
	if err != nil {
		return nil, messageprocessors.ErrInvalidInput.NewWithCause(nil, err)
	}

	env := h.createEnvironment(model)
	env["application_id"] = msg.ApplicationID
	env["device_id"] = msg.DeviceID
	env["dev_eui"] = msg.DevEUI
	env["join_eui"] = msg.JoinEUI
	env["payload"] = m
	env["f_port"] = payload.FPort

	script = fmt.Sprintf(`
		%s
		Encoder(env.payload, env.f_port)
	`, script)

	value, err := h.engine.Run(ctx, script, env)
	if err != nil {
		return nil, err
	}

	if value == nil || reflect.TypeOf(value).Kind() != reflect.Slice {
		return nil, messageprocessors.ErrInvalidOutputType.New(nil)
	}

	slice := reflect.ValueOf(value)
	l := slice.Len()
	payload.FRMPayload = make([]byte, l)
	for i := 0; i < l; i++ {
		val := slice.Index(i).Interface()
		var b int64
		switch i := val.(type) {
		case int:
			b = int64(i)
		case int8:
			b = int64(i)
		case int16:
			b = int64(i)
		case int32:
			b = int64(i)
		case int64:
			b = i
		case uint8:
			b = int64(i)
		case uint16:
			b = int64(i)
		case uint32:
			b = int64(i)
		case uint64:
			b = int64(i)
		default:
			return nil, messageprocessors.ErrInvalidOutput.New(nil)
		}
		if b < 0x00 || b > 0xFF {
			return nil, messageprocessors.ErrInvalidOutputRange.New(errors.Attributes{
				"value": b,
				"low":   0x00,
				"high":  0xFF,
			})
		}
		payload.FRMPayload[i] = byte(b)
	}

	return msg, nil
}

// Decode decodes the message's MAC payload FRMPayload to DecodedPayload using script.
func (h *host) Decode(ctx context.Context, msg *ttnpb.UplinkMessage, model *ttnpb.EndDeviceModel, script string) (*ttnpb.UplinkMessage, error) {
	payload := msg.Payload.GetMACPayload()
	if payload == nil {
		return nil, messageprocessors.ErrNoMACPayload.New(nil)
	}

	env := h.createEnvironment(model)
	env["application_id"] = msg.ApplicationID
	env["device_id"] = msg.DeviceID
	env["dev_eui"] = msg.DevEUI
	env["join_eui"] = msg.JoinEUI
	env["payload"] = payload.FRMPayload
	env["f_port"] = payload.FPort

	script = fmt.Sprintf(`
		%s
		Decoder(env.payload, env.f_port)
	`, script)

	value, err := h.engine.Run(ctx, script, env)
	if err != nil {
		return nil, err
	}

	m, ok := value.(map[string]interface{})
	if !ok {
		return nil, messageprocessors.ErrInvalidOutput.New(nil)
	}

	s, err := gogoproto.Struct(m)
	if err != nil {
		return nil, messageprocessors.ErrInvalidOutput.NewWithCause(nil, err)
	}

	payload.DecodedPayload = s
	return msg, nil
}