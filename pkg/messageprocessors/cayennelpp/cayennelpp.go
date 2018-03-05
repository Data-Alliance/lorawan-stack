// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

// Package cayennelpp contains the CayenneLPP payload formatter message processors.
package cayennelpp

import (
	"bytes"
	"context"

	lpp "github.com/TheThingsNetwork/go-cayenne-lib/cayennelpp"
	"github.com/TheThingsNetwork/ttn/pkg/gogoproto"
	"github.com/TheThingsNetwork/ttn/pkg/messageprocessors"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
)

type host struct {
}

type decodedMap map[string]interface{}

// New creates and returns a new CayenneLPP payload encoder and decoder.
func New() messageprocessors.PayloadEncodeDecoder {
	return &host{}
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

	encoder := lpp.NewEncoder()
	for name, value := range m {
		key, channel, err := parseName(name)
		if err != nil {
			continue
		}
		switch key {
		case valueKey:
			if val, ok := value.(float64); ok {
				encoder.AddPort(channel, float32(val))
			}
		}
	}
	payload.FRMPayload = encoder.Bytes()
	return msg, nil
}

// Decode decodes the message's MAC payload FRMPayload to DecodedPayload using script.
func (h *host) Decode(ctx context.Context, msg *ttnpb.UplinkMessage, model *ttnpb.EndDeviceModel, script string) (*ttnpb.UplinkMessage, error) {
	payload := msg.Payload.GetMACPayload()
	if payload == nil {
		return nil, messageprocessors.ErrNoMACPayload.New(nil)
	}

	decoder := lpp.NewDecoder(bytes.NewBuffer(payload.FRMPayload))
	m := decodedMap(make(map[string]interface{}))
	if err := decoder.DecodeUplink(m); err != nil {
		return nil, messageprocessors.ErrInvalidOutput.NewWithCause(nil, err)
	}

	s, err := gogoproto.Struct(m)
	if err != nil {
		return nil, messageprocessors.ErrInvalidOutputType.NewWithCause(nil, err)
	}

	payload.DecodedPayload = s
	return msg, nil
}

func (d decodedMap) DigitalInput(channel, value uint8) {
	d[formatName(digitalInputKey, channel)] = value
}

func (d decodedMap) DigitalOutput(channel, value uint8) {
	d[formatName(digitalOutputKey, channel)] = value
}

func (d decodedMap) AnalogInput(channel uint8, value float32) {
	d[formatName(analogInputKey, channel)] = value
}

func (d decodedMap) AnalogOutput(channel uint8, value float32) {
	d[formatName(analogOutputKey, channel)] = value
}

func (d decodedMap) Luminosity(channel uint8, value uint16) {
	d[formatName(luminosityKey, channel)] = value
}

func (d decodedMap) Presence(channel, value uint8) {
	d[formatName(presenceKey, channel)] = value
}

func (d decodedMap) Temperature(channel uint8, celsius float32) {
	d[formatName(temperatureKey, channel)] = celsius
}

func (d decodedMap) RelativeHumidity(channel uint8, rh float32) {
	d[formatName(relativeHumidityKey, channel)] = rh
}

func (d decodedMap) Accelerometer(channel uint8, x, y, z float32) {
	d[formatName(accelerometerKey, channel)] = map[string]float32{
		"x": x,
		"y": y,
		"z": z,
	}
}

func (d decodedMap) BarometricPressure(channel uint8, hpa float32) {
	d[formatName(barometricPressureKey, channel)] = hpa
}

func (d decodedMap) Gyrometer(channel uint8, x, y, z float32) {
	d[formatName(gyrometerKey, channel)] = map[string]float32{
		"x": x,
		"y": y,
		"z": z,
	}
}

func (d decodedMap) GPS(channel uint8, latitude, longitude, altitude float32) {
	d[formatName(gpsKey, channel)] = map[string]float32{
		"latitude":  latitude,
		"longitude": longitude,
		"altitude":  altitude,
	}
}