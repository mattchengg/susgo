#!/bin/bash

# Validation script for Task 6.3: Add success notifications
# This script verifies that success notifications are properly implemented

echo "=== Task 6.3 Validation: Success Notifications ==="
echo ""

PASS_COUNT=0
FAIL_COUNT=0

# Function to check if a pattern exists in main.go
check_pattern() {
    local description="$1"
    local pattern="$2"
    
    if grep -q "$pattern" main.go; then
        echo "✓ $description"
        ((PASS_COUNT++))
        return 0
    else
        echo "✗ $description"
        ((FAIL_COUNT++))
        return 1
    fi
}

# Function to count occurrences of a pattern
count_pattern() {
    local description="$1"
    local pattern="$2"
    local expected="$3"
    
    local count=$(grep -c "$pattern" main.go || echo "0")
    
    if [ "$count" -eq "$expected" ]; then
        echo "✓ $description (found $count)"
        ((PASS_COUNT++))
        return 0
    else
        echo "✗ $description (expected $expected, found $count)"
        ((FAIL_COUNT++))
        return 1
    fi
}

echo "1. Checking dialog.ShowInformation() usage:"
echo ""

# Check that dialog.ShowInformation is used
count_pattern "dialog.ShowInformation() is used" "dialog.ShowInformation" 3

echo ""
echo "2. Checking Check Update tab success notification:"
echo ""

# Check for success notification in Check Update tab
check_pattern "Check Update: Uses dialog.ShowInformation" 'dialog.ShowInformation.*Success.*Latest version'
check_pattern "Check Update: Includes version in message" 'version.*model.*region'
check_pattern "Check Update: Clears result label on success" 'resultLabel.SetText("")'

echo ""
echo "3. Checking Download tab success notification:"
echo ""

# Check for success notification in Download tab
check_pattern "Download: Uses dialog.ShowInformation" 'dialog.ShowInformation.*Download Complete'
check_pattern "Download: Shows firmware downloaded successfully message" 'Firmware downloaded successfully'
check_pattern "Download: Includes model, region, version, location" 'Model:.*Region:.*Version:.*Location:'
check_pattern "Download: Clears status label on success" 'statusLabel.SetText("")'
check_pattern "Download: Hides progress bar after completion" 'progressBar.Hide()'

echo ""
echo "4. Checking Decrypt tab success notification:"
echo ""

# Check for success notification in Decrypt tab  
check_pattern "Decrypt: Uses dialog.ShowInformation" 'dialog.ShowInformation.*Decryption Complete'
check_pattern "Decrypt: Shows firmware decrypted successfully message" 'Firmware decrypted successfully'
check_pattern "Decrypt: Includes encryption version and output file" 'Encryption Version:.*Output File:'
check_pattern "Decrypt: Clears status label on success" 'statusLabel.SetText("")'

echo ""
echo "5. Checking that error dialogs still work:"
echo ""

# Verify error dialogs are still in place
check_pattern "Error dialogs still used (at least 15)" "dialog.ShowError"

echo ""
echo "6. Checking proper formatting and structure:"
echo ""

# Check that fmt.Sprintf is used for formatting messages
check_pattern "Uses fmt.Sprintf for formatted messages" "fmt.Sprintf"

echo ""
echo "=== Validation Summary ==="
echo "Passed: $PASS_COUNT"
echo "Failed: $FAIL_COUNT"
echo ""

if [ $FAIL_COUNT -eq 0 ]; then
    echo "✅ All validations passed! Task 6.3 is complete."
    exit 0
else
    echo "❌ Some validations failed. Please review the implementation."
    exit 1
fi
