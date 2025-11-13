package service

import (
	"fmt"
	"log"
	"time"

	"api/internal/libs/logger"
	"github.com/google/uuid"
)

// Fraud detection thresholds
const (
	HIGH_VALUE_THRESHOLD    = 10000.0 // $10,000
	RAPID_DEPLETION_HOURS   = 24
	SUSPICIOUS_USAGE_AMOUNT = 5000.0 // Single transaction over $5,000
)

// detectFraudOnCreation checks for suspicious patterns when a subsidy is created
func (s *SubsidyService) detectFraudOnCreation(customerID, subsidyID uuid.UUID, amount float64, staffID uuid.UUID, ipAddress string) {
	// Alert: High-value subsidy
	if amount >= HIGH_VALUE_THRESHOLD {
		log.Printf("🚨 [FRAUD-ALERT] [HIGH] High-value subsidy created: Subsidy=%s, Customer=%s, Amount=$%.2f, Staff=%s, IP=%s",
			subsidyID, customerID, amount, staffID, ipAddress)

		// Send Slack alert for high-value subsidies
		logger.SendSlackAlertAsync("FRAUD_ALERT",
			fmt.Sprintf("High-value subsidy created: $%.2f", amount),
			map[string]interface{}{
				"Severity":    "HIGH",
				"Type":        "High Value Subsidy",
				"Subsidy ID":  subsidyID.String(),
				"Customer ID": customerID.String(),
				"Amount":      fmt.Sprintf("$%.2f", amount),
				"Threshold":   fmt.Sprintf("$%.2f", HIGH_VALUE_THRESHOLD),
				"Staff ID":    staffID.String(),
				"IP Address":  ipAddress,
				"Timestamp":   time.Now().Format(time.RFC3339),
			})
	}
}

// detectFraudOnUsage checks for suspicious patterns when subsidy is used
func (s *SubsidyService) detectFraudOnUsage(customerID, subsidyID uuid.UUID, usageAmount, remainingBalance float64, createdAt time.Time) {
	// Alert 1: Large single transaction
	if usageAmount >= SUSPICIOUS_USAGE_AMOUNT {
		log.Printf("🔶 [FRAUD-ALERT] [MEDIUM] Large subsidy usage: Subsidy=%s, Customer=%s, Amount=$%.2f",
			subsidyID, customerID, usageAmount)

		logger.SendSlackAlertAsync("FRAUD_ALERT",
			fmt.Sprintf("Large subsidy transaction: $%.2f", usageAmount),
			map[string]interface{}{
				"Severity":    "MEDIUM",
				"Type":        "Large Transaction",
				"Subsidy ID":  subsidyID.String(),
				"Customer ID": customerID.String(),
				"Amount":      fmt.Sprintf("$%.2f", usageAmount),
				"Threshold":   fmt.Sprintf("$%.2f", SUSPICIOUS_USAGE_AMOUNT),
				"Remaining":   fmt.Sprintf("$%.2f", remainingBalance),
				"Timestamp":   time.Now().Format(time.RFC3339),
			})
	}

	// Alert 2: Rapid depletion (subsidy fully used within 24 hours)
	hoursSinceCreation := time.Since(createdAt).Hours()
	if remainingBalance == 0 && hoursSinceCreation <= RAPID_DEPLETION_HOURS {
		log.Printf("🚨 [FRAUD-ALERT] [HIGH] Rapid subsidy depletion: Subsidy=%s, Customer=%s, Depleted in %.1f hours",
			subsidyID, customerID, hoursSinceCreation)

		logger.SendSlackAlertSync("FRAUD_ALERT",
			fmt.Sprintf("Subsidy depleted in %.1f hours (threshold: %d hrs)", hoursSinceCreation, RAPID_DEPLETION_HOURS),
			map[string]interface{}{
				"Severity":      "HIGH",
				"Type":          "Rapid Depletion",
				"Subsidy ID":    subsidyID.String(),
				"Customer ID":   customerID.String(),
				"Hours":         fmt.Sprintf("%.1f", hoursSinceCreation),
				"Threshold":     fmt.Sprintf("%d hours", RAPID_DEPLETION_HOURS),
				"Final Amount":  fmt.Sprintf("$%.2f", usageAmount),
				"Action":        "Review customer subsidy usage patterns",
				"Timestamp":     time.Now().Format(time.RFC3339),
			})
	}
}
