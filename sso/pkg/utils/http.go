package utils

import (
	"context"
	"crypto/sha256"
	"fmt"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"net/http"
	"sort"
	"strings"
)

func convertMetadataToHTTPHeader(md metadata.MD) http.Header {
	header := make(http.Header)
	for key, values := range md {
		for _, value := range values {
			header.Add(key, value)
		}
	}
	return header
}

func GetFingerprint(md metadata.MD) string {
	headers := convertMetadataToHTTPHeader(md)

	var headerList []string
	for key, values := range headers {
		if shouldIncludeHeader(key) { // Проверяем, нужен ли нам этот заголовок
			headerList = append(headerList, key+": "+strings.Join(values, ","))
		}
	}
	sort.Strings(headerList)
	sortedHeaders := strings.Join(headerList, ",")
	hash := sha256.Sum256([]byte(sortedHeaders))
	return fmt.Sprintf("%x", hash)
}

func GetGRPCClientIP(ctx context.Context, md metadata.MD) (string, error) {
	// Получаем заголовок X-Forwarded-For
	forwardedFor := md.Get("x-forwarded-for")
	var clientIP string
	if len(forwardedFor) > 0 {
		// X-Forwarded-For может содержать несколько IP-адресов, разделенных запятыми
		clientIP = strings.Split(forwardedFor[0], ",")[0]
	}

	// Если X-Forwarded-For отсутствует, пытаемся получить IP через peer.Peer
	if clientIP == "" {
		p, ok := peer.FromContext(ctx)
		if ok {
			clientIP = p.Addr.String()
		}
	}

	if clientIP == "" {
		return "", fmt.Errorf("failed to get client IP")
	}

	fmt.Println("Client IP:", clientIP)
	return clientIP, nil
}

func shouldIncludeHeader(key string) bool {
	switch key {
	case "User-Agent", "Content-Type", ":authority", "Grpc-Accept-Encoding":
		return true
	default:
		return false
	}
}
