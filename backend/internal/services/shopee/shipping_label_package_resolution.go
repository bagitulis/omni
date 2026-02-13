package shopee

import (
	shopeePkg "github.com/omni/backend/pkg/shopee"
)

func resolvePackageNumber(client shippingClient, orderSN, packageNumber string, trackingResp *shopeePkg.GetTrackingNumberResponse) (string, string) {
	if packageNumber != "" {
		return packageNumber, ""
	}

	trackingNumber := ""
	if trackingResp != nil {
		trackingNumber = trackingResp.Response.TrackingNumber
	}

	resolved, docError := resolvePackageNumberFromDocumentData(client, orderSN)
	if resolved != "" {
		return resolved, ""
	}

	resolved, packageError := resolvePackageNumberFromPackageAPI(client, orderSN, trackingNumber)
	if resolved != "" {
		return resolved, ""
	}

	return "", firstNonEmpty(docError, packageError)
}

func resolvePackageNumberFromDocumentData(client shippingClient, orderSN string) (string, string) {
	resp, err := client.GetShippingDocumentDataInfo(orderSN, "")
	if err != nil {
		return "", err.Error()
	}

	if resp == nil {
		return "", ""
	}

	if rawError := formatShopeeError(resp.Error, resp.Message); rawError != "" {
		return "", rawError
	}

	return resp.Response.PackageNumber, ""
}

func resolvePackageNumberFromPackageAPI(client shippingClient, orderSN, trackingNumber string) (string, string) {
	packageStatuses := []int{3, 2, 1, 0}

	for _, status := range packageStatuses {
		cursor := ""
		for page := 0; page < 3; page++ {
			resp, err := client.SearchPackageList(status, cursor, 100)
			if err != nil {
				return "", err.Error()
			}
			if resp == nil {
				break
			}

			for _, pkg := range resp.Response.PackagesList {
				if pkg.OrderSN == orderSN && pkg.PackageNumber != "" {
					return pkg.PackageNumber, ""
				}
			}

			if trackingNumber != "" {
				packageNumbers := collectPackageNumbers(resp.Response.PackagesList)
				if len(packageNumbers) > 0 {
					resolved, err := resolvePackageByDetailMatch(client, packageNumbers, orderSN, trackingNumber)
					if err != nil {
						return "", err.Error()
					}
					if resolved != "" {
						return resolved, ""
					}
				}
			}

			if !resp.Response.Pagination.More || resp.Response.Pagination.NextCursor == "" {
				break
			}
			cursor = resp.Response.Pagination.NextCursor
		}
	}

	return "", ""
}

func collectPackageNumbers(packages []shopeePkg.PackageBasic) []string {
	seen := make(map[string]struct{}, len(packages))
	result := make([]string, 0, len(packages))

	for _, pkg := range packages {
		if pkg.PackageNumber == "" {
			continue
		}
		if _, exists := seen[pkg.PackageNumber]; exists {
			continue
		}
		seen[pkg.PackageNumber] = struct{}{}
		result = append(result, pkg.PackageNumber)
	}

	return result
}

func resolvePackageByDetailMatch(client shippingClient, packageNumbers []string, orderSN, trackingNumber string) (string, error) {
	for _, batch := range chunkStrings(packageNumbers, 50) {
		detailResp, err := client.GetPackageDetail(batch)
		if err != nil {
			return "", err
		}
		if detailResp == nil {
			continue
		}

		for _, detail := range detailResp.Response.PackageList {
			if detail.OrderSN == orderSN && detail.TrackingNumber == trackingNumber && detail.PackageNumber != "" {
				return detail.PackageNumber, nil
			}
		}
	}

	return "", nil
}

func chunkStrings(values []string, size int) [][]string {
	if size <= 0 || len(values) == 0 {
		return nil
	}

	chunks := make([][]string, 0, (len(values)+size-1)/size)
	for i := 0; i < len(values); i += size {
		end := i + size
		if end > len(values) {
			end = len(values)
		}
		chunks = append(chunks, values[i:end])
	}

	return chunks
}
