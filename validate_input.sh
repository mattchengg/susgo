#!/bin/bash
# Validation script for Task 6.1: Add Input Validation

echo "=== Task 6.1 Input Validation - Validation ==="
echo ""

echo "✓ Checking validation helper functions..."
grep -q "func validateModel(model string) error" main.go && echo "  ✓ validateModel() function defined" || echo "  ✗ Missing validateModel() function"
grep -q "func validateRegion(region string) error" main.go && echo "  ✓ validateRegion() function defined" || echo "  ✗ Missing validateRegion() function"
grep -q "func validateIMEI(imei string) error" main.go && echo "  ✓ validateIMEI() function defined" || echo "  ✗ Missing validateIMEI() function"

echo ""
echo "✓ Checking validateModel requirements..."
grep -q 'strings.HasPrefix.*"SM-"' main.go && echo "  ✓ Checks model starts with SM-" || echo "  ✗ Missing SM- prefix check"
grep -q 'model == ""' main.go && echo "  ✓ Checks model is non-empty" || echo "  ✗ Missing empty check"
grep -q 'len(model) < 5' main.go && echo "  ✓ Checks minimum model length" || echo "  ✗ Missing length check"

echo ""
echo "✓ Checking validateRegion requirements..."
grep -q 'region == ""' main.go && echo "  ✓ Checks region is non-empty" || echo "  ✗ Missing empty check"
grep -q 'len(region) < 2 || len(region) > 4' main.go && echo "  ✓ Checks region is 2-4 characters" || echo "  ✗ Missing length check"
grep -q "char < 'A' || char > 'Z'" main.go && echo "  ✓ Checks region contains only letters" || echo "  ✗ Missing letter check"

echo ""
echo "✓ Checking validateIMEI requirements..."
grep -q 'imei == ""' main.go && echo "  ✓ Checks IMEI is non-empty" || echo "  ✗ Missing empty check"
grep -q 'len(imei) != 8 && len(imei) != 15' main.go && echo "  ✓ Checks IMEI is 8 or 15 digits" || echo "  ✗ Missing length check"
grep -q 'regexp.MatchString.*\\d' main.go && echo "  ✓ Checks IMEI contains only digits" || echo "  ✗ Missing digit check"

echo ""
echo "✓ Checking validation usage in Check Update Tab..."
grep -q 'validateModel(model)' main.go && echo "  ✓ Uses validateModel() in Check Update" || echo "  ✗ Check Update tab doesn't use validateModel()"
grep -q 'validateRegion(region)' main.go && echo "  ✓ Uses validateRegion() in Check Update" || echo "  ✗ Check Update tab doesn't use validateRegion()"

echo ""
echo "✓ Checking validation usage in Download Tab..."
# Count validateModel usage - should be at least 2 (Check Update + Download)
MODEL_COUNT=$(grep -c 'validateModel(model)' main.go)
if [ "$MODEL_COUNT" -ge 2 ]; then
    echo "  ✓ Uses validateModel() in Download tab"
else
    echo "  ✗ Download tab doesn't use validateModel()"
fi

REGION_COUNT=$(grep -c 'validateRegion(region)' main.go)
if [ "$REGION_COUNT" -ge 2 ]; then
    echo "  ✓ Uses validateRegion() in Download tab"
else
    echo "  ✗ Download tab doesn't use validateRegion()"
fi

grep -q 'validateIMEI(imei)' main.go && echo "  ✓ Uses validateIMEI() in Download tab" || echo "  ✗ Download tab doesn't use validateIMEI()"

echo ""
echo "✓ Checking validation usage in Decrypt Tab..."
# Count should be at least 3 for all tabs
MODEL_COUNT=$(grep -c 'validateModel(model)' main.go)
if [ "$MODEL_COUNT" -ge 3 ]; then
    echo "  ✓ Uses validateModel() in Decrypt tab"
else
    echo "  ✗ Decrypt tab doesn't use validateModel()"
fi

REGION_COUNT=$(grep -c 'validateRegion(region)' main.go)
if [ "$REGION_COUNT" -ge 3 ]; then
    echo "  ✓ Uses validateRegion() in Decrypt tab"
else
    echo "  ✗ Decrypt tab doesn't use validateRegion()"
fi

IMEI_COUNT=$(grep -c 'validateIMEI(imei)' main.go)
if [ "$IMEI_COUNT" -ge 2 ]; then
    echo "  ✓ Uses validateIMEI() in Decrypt tab"
else
    echo "  ✗ Decrypt tab doesn't use validateIMEI()"
fi

echo ""
echo "✓ Checking error message clarity..."
grep -q 'err.Error()' main.go && echo "  ✓ Error messages are displayed to user" || echo "  ✗ Missing error display"

echo ""
echo "✓ Checking regexp import..."
grep -q '"regexp"' main.go && echo "  ✓ regexp package imported" || echo "  ✗ Missing regexp import"

echo ""
echo "=== Validation Complete ==="
