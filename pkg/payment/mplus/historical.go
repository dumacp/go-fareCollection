package mplus

import "time"

type historicalUse struct {
	index           int
	fareID          uint
	timeTransaction time.Time
	itineraryID     uint
	deviceID        uint
}
type historicalRecharge struct {
	// FareID          int
	index           int
	timeTransaction time.Time
	rechargedID     uint
	typeTransaction uint
	// ItineraryID     int
	value    int
	deviceID uint
}

func (h *historicalUse) FareID() uint {
	return h.fareID
}

func (h *historicalUse) Index() int {
	return h.index
}

func (h *historicalUse) TimeTransaction() time.Time {
	return h.timeTransaction
}

func (h *historicalUse) ItineraryID() uint {
	return h.itineraryID
}

func (h *historicalUse) DeviceID() uint {
	return h.deviceID
}

func (h *historicalUse) SetFareID(fareID uint) {
	h.fareID = fareID
}

func (h *historicalUse) SetIndex(index int) {
	h.index = index
}

func (h *historicalUse) SetTimeTransaction(timeTransaction time.Time) {
	h.timeTransaction = timeTransaction
}

func (h *historicalUse) SetItineraryID(itineraryID uint) {
	h.itineraryID = itineraryID
}

func (h *historicalUse) SetDeviceID(deviceID uint) {
	h.deviceID = deviceID
}

func (h *historicalRecharge) Index() int {
	return h.index
}

func (h *historicalRecharge) SetIndex(index int) {
	h.index = index
}

func (h *historicalRecharge) DeviceID() uint {
	return h.deviceID
}

func (h *historicalRecharge) Value() int {
	return h.value
}

func (h *historicalRecharge) TimeTransaction() time.Time {
	return h.timeTransaction
}

func (h *historicalRecharge) ConsecutiveID() uint {
	return h.rechargedID
}

func (h *historicalRecharge) TypeTransaction() uint {
	return h.typeTransaction
}

func (h *historicalRecharge) SetDeviceID(deviceID uint) {
	h.deviceID = deviceID
}

func (h *historicalRecharge) SetValue(value int) {
	h.value = value
}

func (h *historicalRecharge) SetTimeTransaction(timeTransaction time.Time) {
	h.timeTransaction = timeTransaction
}

func (h *historicalRecharge) SetConsecutiveID(consecutive uint) {
	h.rechargedID = consecutive
}

func (h *historicalRecharge) SetTypeTransaction(typeTransaction uint) {
	h.typeTransaction = typeTransaction
}
