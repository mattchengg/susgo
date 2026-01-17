#!/bin/bash

echo "==================================="
echo "Validating Error Dialog Implementation"
echo "==================================="
echo ""

PASSED=0
FAILED=0

# Test 1: Check if makeCheckUpdateTab accepts window parameter
echo "Test 1: makeCheckUpdateTab accepts window parameter"
if grep -q "^func makeCheckUpdateTab(window fyne.Window)" main.go; then
    echo "✓ PASS: makeCheckUpdateTab has window parameter"
    ((PASSED++))
else
    echo "✗ FAIL: makeCheckUpdateTab doesn't have window parameter"
    ((FAILED++))
fi

# Test 2: Check if makeDownloadTab accepts window parameter
echo "Test 2: makeDownloadTab accepts window parameter"
if grep -q "^func makeDownloadTab(window fyne.Window)" main.go; then
    echo "✓ PASS: makeDownloadTab has window parameter"
    ((PASSED++))
else
    echo "✗ FAIL: makeDownloadTab doesn't have window parameter"
    ((FAILED++))
fi

# Test 3: Check if makeDecryptTab accepts window parameter
echo "Test 3: makeDecryptTab accepts window parameter"
if grep -q "^func makeDecryptTab(window fyne.Window)" main.go; then
    echo "✓ PASS: makeDecryptTab has window parameter"
    ((PASSED++))
else
    echo "✗ FAIL: makeDecryptTab doesn't have window parameter"
    ((FAILED++))
fi

# Test 4: Check makeCheckUpdateTab uses dialog.ShowError for validation errors
echo "Test 4: makeCheckUpdateTab uses dialog.ShowError for validation"
if grep -A 2 "validateModel(model)" main.go | grep -A 1 "makeCheckUpdateTab" | grep -q "dialog.ShowError"; then
    echo "✓ PASS: makeCheckUpdateTab uses dialog.ShowError for model validation"
    ((PASSED++))
else
    echo "✗ FAIL: makeCheckUpdateTab doesn't use dialog.ShowError for model validation"
    ((FAILED++))
fi

# Test 5: Check makeCheckUpdateTab uses dialog.ShowError for API errors
echo "Test 5: makeCheckUpdateTab uses dialog.ShowError for API errors"
if grep -A 3 "getLatestVersion(model, region)" main.go | grep -q "dialog.ShowError(err, window)"; then
    echo "✓ PASS: makeCheckUpdateTab uses dialog.ShowError for API errors"
    ((PASSED++))
else
    echo "✗ FAIL: makeCheckUpdateTab doesn't use dialog.ShowError for API errors"
    ((FAILED++))
fi

# Test 6: Check makeDownloadTab uses dialog.ShowError for validation
echo "Test 6: makeDownloadTab uses dialog.ShowError for validation"
if grep -A 20 "func makeDownloadTab" main.go | grep -c "dialog.ShowError" | grep -q "[3-9]"; then
    echo "✓ PASS: makeDownloadTab uses dialog.ShowError for validation (multiple times)"
    ((PASSED++))
else
    echo "✗ FAIL: makeDownloadTab doesn't use dialog.ShowError enough times"
    ((FAILED++))
fi

# Test 7: Check makeDownloadTab uses dialog.ShowError for download errors
echo "Test 7: makeDownloadTab uses dialog.ShowError for download errors"
if grep -A 10 "downloadFirmware(model, region" main.go | grep "if err != nil" -A 3 | grep -q "dialog.ShowError(err, window)"; then
    echo "✓ PASS: makeDownloadTab uses dialog.ShowError for download errors"
    ((PASSED++))
else
    echo "✗ FAIL: makeDownloadTab doesn't use dialog.ShowError for download errors"
    ((FAILED++))
fi

# Test 8: Check makeDecryptTab uses dialog.ShowError for validation
echo "Test 8: makeDecryptTab uses dialog.ShowError for validation"
if grep -A 80 "func makeDecryptTab" main.go | grep -c "dialog.ShowError" | grep -q "[5-9]"; then
    echo "✓ PASS: makeDecryptTab uses dialog.ShowError for validation (multiple times)"
    ((PASSED++))
else
    echo "✗ FAIL: makeDecryptTab doesn't use dialog.ShowError enough times"
    ((FAILED++))
fi

# Test 9: Check makeDecryptTab uses dialog.ShowError for decryption errors
echo "Test 9: makeDecryptTab uses dialog.ShowError for decryption errors"
if grep -A 10 "decryptFirmwareGUI(model, region" main.go | grep "if err != nil" -A 3 | grep -q "dialog.ShowError(err, window)"; then
    echo "✓ PASS: makeDecryptTab uses dialog.ShowError for decryption errors"
    ((PASSED++))
else
    echo "✗ FAIL: makeDecryptTab doesn't use dialog.ShowError for decryption errors"
    ((FAILED++))
fi

# Test 10: Check main() passes window to all tabs
echo "Test 10: main() passes window to all tabs"
if grep "makeCheckUpdateTab(myWindow)" main.go && \
   grep "makeDownloadTab(myWindow)" main.go && \
   grep "makeDecryptTab(myWindow)" main.go; then
    echo "✓ PASS: main() passes window to all tab functions"
    ((PASSED++))
else
    echo "✗ FAIL: main() doesn't pass window to all tab functions"
    ((FAILED++))
fi

# Test 11: Check NO error status labels in makeCheckUpdateTab validation
echo "Test 11: No error status labels for validation in makeCheckUpdateTab"
if ! grep -A 3 "validateModel(model)" main.go | grep -A 1 "makeCheckUpdateTab" | grep "resultLabel.SetText.*Error"; then
    echo "✓ PASS: makeCheckUpdateTab doesn't use status labels for validation errors"
    ((PASSED++))
else
    echo "✗ FAIL: makeCheckUpdateTab still uses status labels for validation errors"
    ((FAILED++))
fi

# Test 12: Check NO error status labels in makeDownloadTab validation
echo "Test 12: No error status labels for validation in makeDownloadTab"
if ! grep -A 50 "func makeDownloadTab" main.go | grep "validateModel\|validateRegion\|validateIMEI" -A 2 | grep "statusLabel.SetText.*Error"; then
    echo "✓ PASS: makeDownloadTab doesn't use status labels for validation errors"
    ((PASSED++))
else
    echo "✗ FAIL: makeDownloadTab still uses status labels for validation errors"
    ((FAILED++))
fi

# Test 13: Check NO error status labels in makeDecryptTab validation
echo "Test 13: No error status labels for validation in makeDecryptTab"
if ! grep -A 80 "func makeDecryptTab" main.go | grep "validateModel\|validateRegion\|validateIMEI" -A 2 | grep "statusLabel.SetText.*Error"; then
    echo "✓ PASS: makeDecryptTab doesn't use status labels for validation errors"
    ((PASSED++))
else
    echo "✗ FAIL: makeDecryptTab still uses status labels for validation errors"
    ((FAILED++))
fi

# Test 14: Check status labels are used for progress messages (not errors)
echo "Test 14: Status labels used for progress messages in makeDownloadTab"
if grep "statusLabel.SetText.*Initializing download" main.go; then
    echo "✓ PASS: Status labels used for progress messages"
    ((PASSED++))
else
    echo "✗ FAIL: Status labels not used for progress messages"
    ((FAILED++))
fi

# Test 15: Check status labels are used for progress messages in makeDecryptTab
echo "Test 15: Status labels used for progress messages in makeDecryptTab"
if grep "statusLabel.SetText.*Starting decryption" main.go; then
    echo "✓ PASS: Status labels used for progress messages in makeDecryptTab"
    ((PASSED++))
else
    echo "✗ FAIL: Status labels not used for progress messages in makeDecryptTab"
    ((FAILED++))
fi

echo ""
echo "==================================="
echo "Results: $PASSED passed, $FAILED failed"
echo "==================================="

if [ $FAILED -eq 0 ]; then
    echo "✅ All tests passed!"
    exit 0
else
    echo "❌ Some tests failed"
    exit 1
fi
