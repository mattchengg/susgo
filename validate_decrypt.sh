#!/bin/bash
# Validation script for Task 5.3: Decrypt Logic with GUI

echo "=== Task 5.3 Validation ==="
echo ""

echo "✓ Checking decryptFirmwareGUI function signature..."
grep -q "func decryptFirmwareGUI.*ProgressReporter" main.go && echo "  ✓ Uses ProgressReporter interface" || echo "  ✗ Missing ProgressReporter parameter"

echo ""
echo "✓ Checking parseIMEI validation..."
grep -q "parseIMEI(imei" main.go && echo "  ✓ Calls parseIMEI()" || echo "  ✗ Missing parseIMEI() call"

echo ""
echo "✓ Checking key generation..."
grep -q 'getV2Key(version, model, region)' main.go && echo "  ✓ Uses getV2Key() for V2" || echo "  ✗ Missing getV2Key()"
grep -q 'getV4Key(version, model, region, effectiveIMEI)' main.go && echo "  ✓ Uses getV4Key() for V4" || echo "  ✗ Missing getV4Key()"

echo ""
echo "✓ Checking decryptFirmwareWithProgress call..."
grep -q "decryptFirmwareWithProgress.*progress)" main.go && echo "  ✓ Calls decryptFirmwareWithProgress() with progress" || echo "  ✗ Missing decryptFirmwareWithProgress() call"

echo ""
echo "✓ Checking progress bar in makeDecryptTab..."
grep -q "progressBar := widget.NewProgressBar()" main.go && echo "  ✓ Progress bar widget created" || echo "  ✗ Missing progress bar widget"
grep -q "progressBar.Show()" main.go && echo "  ✓ Progress bar shown during operation" || echo "  ✗ Progress bar not shown"
grep -q "progressBar.Hide()" main.go && echo "  ✓ Progress bar hidden after operation" || echo "  ✗ Progress bar not hidden"

echo ""
echo "✓ Checking button state management..."
grep -q "decryptButton.Disable()" main.go && echo "  ✓ Button disabled during operation" || echo "  ✗ Button not disabled"
grep -q "decryptButton.Enable()" main.go && echo "  ✓ Button re-enabled after operation" || echo "  ✗ Button not re-enabled"

echo ""
echo "✓ Checking error handling..."
grep -q 'return fmt.Errorf.*invalid IMEI' main.go && echo "  ✓ IMEI validation error handled" || echo "  ✗ Missing IMEI error"
grep -q 'return fmt.Errorf.*failed to generate V4 key' main.go && echo "  ✓ Key generation error handled" || echo "  ✗ Missing key error"
grep -q 'return fmt.Errorf.*decryption failed' main.go && echo "  ✓ Decryption error handled" || echo "  ✗ Missing decryption error"

echo ""
echo "✓ Checking progress.Finish() call..."
grep -q "progress.Finish()" main.go && echo "  ✓ Calls progress.Finish()" || echo "  ✗ Missing progress.Finish()"

echo ""
echo "✓ Checking decryptFirmwareWithProgress implementation in crypt.go..."
grep -q "func decryptFirmwareWithProgress" crypt.go && echo "  ✓ Function defined in crypt.go" || echo "  ✗ Function not defined"
grep -q "progress.SetTotal(length)" crypt.go && echo "  ✓ Sets total progress" || echo "  ✗ Missing SetTotal()"
grep -q "progress.SetCurrent(processed)" crypt.go && echo "  ✓ Updates current progress" || echo "  ✗ Missing SetCurrent()"
grep -q "progress.SetStatus" crypt.go && echo "  ✓ Sets status messages" || echo "  ✗ Missing SetStatus()"

echo ""
echo "=== Validation Complete ==="
