package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var (
	randSource = rand.NewSource(time.Now().UnixNano())
	randRange  = rand.New(randSource)
)

// getTestParcel возвращает тестовую посылку
func getTestParcel() Parcel {
	return Parcel{
		Client:    1000,
		Status:    ParcelStatusRegistered,
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

// TestAddGetDelete проверяет добавление, получение и удаление посылки
func TestAddGetDelete(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err, "failed to open database")
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	id, err := store.Add(parcel)
	require.NoError(t, err, "failed to add parcel")
	parcel.Number = id

	// get
	retrievedParcel, err := store.Get(parcel.Number)
	require.NoError(t, err, "failed to get parcel")
	require.Equal(t, parcel, retrievedParcel, "retrieved parcel does not match original")

	// delete
	err = store.Delete(parcel.Number)
	require.NoError(t, err, "failed to delete parcel")

	// Check that the parcel cannot be retrieved anymore
	_, err = store.Get(parcel.Number)
	require.Error(t, err, "expected error when getting deleted parcel")
}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err, "failed to open database")
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	id, err := store.Add(parcel)
	require.NoError(t, err, "failed to add parcel")
	parcel.Number = id

	// set address
	newAddress := "new test address"
	err = store.SetAddress(parcel.Number, newAddress)
	require.NoError(t, err, "failed to set address")

	// check
	retrievedParcel, err := store.Get(parcel.Number)
	require.NoError(t, err, "failed to get parcel after updating address")
	require.Equal(t, newAddress, retrievedParcel.Address, "address was not updated correctly")
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err, "failed to open database")
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	id, err := store.Add(parcel)
	require.NoError(t, err, "failed to add parcel")
	parcel.Number = id

	// set status
	err = store.SetStatus(parcel.Number, ParcelStatusSent)
	require.NoError(t, err, "failed to set status")

	// check
	retrievedParcel, err := store.Get(parcel.Number)
	require.NoError(t, err, "failed to get parcel after updating status")
	require.Equal(t, ParcelStatusSent, retrievedParcel.Status, "status was not updated correctly")
}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err, "failed to open database")
	defer db.Close()

	store := NewParcelStore(db)
	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}
	parcelMap := map[int]Parcel{}

	// задаём всем посылкам один и тот же идентификатор клиента
	client := randRange.Intn(10_000_000)
	for i := range parcels {
		parcels[i].Client = client
	}

	// add
	for i := 0; i < len(parcels); i++ {
		id, err := store.Add(parcels[i])
		require.NoError(t, err, "failed to add parcel")
		parcels[i].Number = id
		parcelMap[id] = parcels[i]
	}

	// get by client
	storedParcels, err := store.GetByClient(client)
	require.NoError(t, err, "failed to get parcels by client")
	require.Equal(t, len(parcels), len(storedParcels), "number of retrieved parcels does not match added parcels")

	// check
	for _, parcel := range storedParcels {
		originalParcel, exists := parcelMap[parcel.Number]
		require.True(t, exists, "retrieved parcel not found in parcelMap")
		require.Equal(t, originalParcel, parcel, "retrieved parcel does not match original")
	}
}
