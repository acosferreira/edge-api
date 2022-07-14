package models

import (
	"errors"
	"testing"

	"github.com/bxcodec/faker/v3"
	"github.com/redhatinsights/edge-api/pkg/db"
)

func TestGroupValidateRequest(t *testing.T) {
	testScenarios := []struct {
		name     string
		group    *DeviceGroup
		expected error
	}{
		{name: "Empty name", group: &DeviceGroup{OrgID: "222222", Account: "111111", Type: "static"}, expected: errors.New(DeviceGroupNameEmptyErrorMessage)},
		{name: "Invalid type", group: &DeviceGroup{Name: "test_group", OrgID: "222222", Account: "111111", Type: "invalid type"}, expected: errors.New(DeviceGroupTypeInvalidErrorMessage)},
		{name: "Invalid name", group: &DeviceGroup{Name: "** test group", OrgID: "222222", Account: "111111", Type: DeviceGroupTypeDefault}, expected: errors.New(DeviceGroupNameInvalidErrorMessage)},
		{name: "Empty account", group: &DeviceGroup{Name: "test_group", Type: "static", OrgID: "222222"}, expected: errors.New(DeviceGroupAccountEmptyErrorMessage)},
		{name: "Valid DeviceGroup", group: &DeviceGroup{Name: "test_group", Account: "111111", OrgID: "222222", Type: DeviceGroupTypeDefault}, expected: nil},
	}

	for _, testScenario := range testScenarios {
		err := testScenario.group.ValidateRequest()
		if err == nil && testScenario.expected != nil {
			t.Errorf("Test %q was supposed to fail but passed successfully", testScenario.name)
		}
		if err != nil && testScenario.expected == nil {
			t.Errorf("Test %q was supposed to pass but failed: %s", testScenario.name, err)
		}
		if err != nil && testScenario.expected != nil && err.Error() != testScenario.expected.Error() {
			t.Errorf("Test %q: expected to fail on %q but got %q", testScenario.name, testScenario.expected, err)
		}
	}
}

func TestGroupCreateUpdateConstraint(t *testing.T) {
	groupInitialAccount := "111111"
	groupInitialName := "test_group"
	groupInitialType := DeviceGroupTypeDynamic
	groupInitialOrgID := "333333"
	groupNewAccount := "222222"
	groupNewOrgID := "444444"

	groupNewType := DeviceGroupTypeStatic
	groupNewName := "new_test_group"

	group := DeviceGroup{Name: groupInitialName, OrgID: groupInitialOrgID, Account: groupInitialAccount, Type: groupInitialType}

	err := group.ValidateRequest()
	if err != nil {
		t.Errorf("Failed to pass validation, Error: %q", err)
	}

	result := db.DB.Create(&group)
	if result.Error != nil {
		t.Errorf("Failed to create DeviceGroup: %q", result.Error)
	}

	var savedGroup DeviceGroup
	result = db.DB.First(&savedGroup, group.ID)
	if result.Error != nil {
		t.Errorf("Failed to retreive the created DeviceGroup: %q", result.Error)
	}

	savedGroup.Account = groupNewAccount
	savedGroup.OrgID = groupNewOrgID
	savedGroup.Type = groupNewType
	savedGroup.Name = groupNewName

	result = db.DB.Save(&savedGroup)
	if result.Error != nil {
		t.Errorf("Failed to save the created DeviceGroup: %q", result.Error)
	}

	var updatedGroup DeviceGroup
	result = db.DB.First(&updatedGroup, group.ID)
	if result.Error != nil {
		t.Errorf("Failed to retreive the updated DeviceGroup: %q", result.Error)
	}
	// The group Account should not be updated
	if updatedGroup.Account != groupInitialAccount {
		t.Errorf("The group Account has been updated expected: %q  but found %q", groupInitialAccount, updatedGroup.Account)
	}
	// The group Type should not be updated
	if updatedGroup.Type != groupInitialType {
		t.Errorf("The group Type has been updated expected: %q  but found %q", groupInitialAccount, updatedGroup.Type)
	}
	// The DeviceGroup Name has to be updated
	if updatedGroup.Name != groupNewName {
		t.Errorf("Failed to update group name expected: %q but found: %q", groupNewName, updatedGroup.Name)
	}
}

func TestBeforeDelete(t *testing.T) {
	orgID := faker.UUIDHyphenated()
	account := faker.UUIDHyphenated()
	deviceGroupName := faker.Name()
	devices := []Device{
		{
			Name:    faker.Name(),
			UUID:    faker.UUIDHyphenated(),
			Account: account,
			OrgID:   orgID,
		},
		{
			Name:    faker.Name(),
			UUID:    faker.UUIDHyphenated(),
			Account: account,
			OrgID:   orgID,
		},
	}
	deviceGroup := &DeviceGroup{
		Name:    deviceGroupName,
		Type:    DeviceGroupTypeDefault,
		Account: account,
		OrgID:   orgID,
		Devices: devices,
	}
	// Create the DeviceGroup
	err := db.DB.Create(&deviceGroup).Error
	if err != nil {
		t.Error("Error saving device group to DB")
	}
	// Get the device group from DB
	err = db.DB.Where("name = ?", deviceGroupName).Find(&deviceGroup).Error
	if err != nil {
		t.Error("Error retrieving device group from DB")
	}
	if len(deviceGroup.Devices) != 2 {
		t.Errorf("Expected 2 devices but found %d: %v", len(deviceGroup.Devices), deviceGroup.Devices)
	}
	// BeforeDelete the DeviceGroup should delete the Devices and not the DeviceGroup
	err = deviceGroup.BeforeDelete(db.DB)
	if err != nil {
		t.Error("Error running BeforeDelete")
	}
	// Get the device group from DB
	err = db.DB.Where("name = ?", deviceGroupName).Find(&deviceGroup).Error
	if err != nil {
		t.Error("Error retrieving device group from DB")
	}
	if len(deviceGroup.Devices) != 0 {
		t.Errorf("Expected 0 devices but found %d: %v", len(deviceGroup.Devices), deviceGroup.Devices)
	}
}

func TestDeviceGroupsBeforeCreate(t *testing.T) {
	orgID := faker.UUIDHyphenated()
	account := faker.UUIDHyphenated()
	deviceGroupNameWithOrgID := faker.Name()
	devices := []Device{
		{
			Name:    faker.Name(),
			UUID:    faker.UUIDHyphenated(),
			OrgID:   orgID,
			Account: account,
		},
	}

	deviceGroupWitOrgID := &DeviceGroup{
		Name:    deviceGroupNameWithOrgID,
		Type:    DeviceGroupTypeDefault,
		OrgID:   orgID,
		Account: account,
		Devices: devices,
	}
	// BeforeCreate make sure DeviceGroup has to orgID
	err := deviceGroupWitOrgID.BeforeCreate(db.DB)
	if err != nil {
		t.Error("Error running BeforeCreate")
	}
}
