package main

import (
	"fmt"
	"math/rand"
	"strings"
)

func luhnChecksum(imei string) int {
	imei += "0"
	parity := len(imei) % 2
	s := 0
	for idx, char := range imei {
		d := int(char - '0')
		if idx%2 == parity {
			d *= 2
			if d > 9 {
				d -= 9
			}
		}
		s += d
	}
	return (10 - (s % 10)) % 10
}

func generateIMEINumbers(tac string, numGenerated int) []string {
	imeiNumbers := make([]string, 0, numGenerated)

	firstDigitChoices := []int{0, 5, 7}
	thirdDigitChoices := []int{0, 1, 3, 5, 6, 7}

	for i := 0; i < numGenerated; i++ {
		rngFirst := firstDigitChoices[rand.Intn(len(firstDigitChoices))]
		rngSecond := rand.Intn(6) + 4 // 4-9
		rngThird := thirdDigitChoices[rand.Intn(len(thirdDigitChoices))]
		rngFourth := rand.Intn(10)
		rngFifthSixth := rand.Intn(100)

		tacRng := fmt.Sprintf("%s%d%d%d%d%02d", tac, rngFirst, rngSecond, rngThird, rngFourth, rngFifthSixth)
		luhnDigit := luhnChecksum(tacRng)
		imei := fmt.Sprintf("%s%d", tacRng, luhnDigit)

		imeiNumbers = append(imeiNumbers, imei)
	}

	return imeiNumbers
}

func generateRandomIMEI(tac string) string {
	if len(tac) == 15 {
		return tac
	} else if len(tac) == 8 {
		imeis := generateIMEINumbers(tac, 1)
		if len(imeis) > 0 {
			return imeis[0]
		}
	}
	return ""
}

func validateAndGenerateIMEI(tac, model, region string) (string, error) {
	if len(tac) == 15 {
		return tac, nil
	}

	if len(tac) != 8 {
		return "", fmt.Errorf("invalid IMEI length: please provide 8 or 15 digits")
	}

	for attempt := 1; attempt <= 5; attempt++ {
		imei := generateRandomIMEI(tac)
		client := NewFUSClient()

		fwVer, err := getLatestVersion(model, region)
		if err != nil {
			return "", err
		}

		req := binaryInform(fwVer, model, region, imei, client.Nonce)
		resp, err := client.MakeReq("NF_DownloadBinaryInform.do", req)
		if err != nil {
			fmt.Printf("Attempt %d: Error during validation: %v\n", attempt, err)
			continue
		}

		if strings.Contains(resp, "<Status>200</Status>") {
			fmt.Printf("Attempt %d: Valid IMEI Found: %s\n", attempt, imei)
			return imei, nil
		}

		fmt.Printf("Attempt %d: IMEI %s is invalid\n", attempt, imei)
	}

	return "", fmt.Errorf("unable to find a valid IMEI after 5 tries")
}
