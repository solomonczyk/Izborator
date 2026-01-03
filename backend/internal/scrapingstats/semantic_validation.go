package scrapingstats

import "github.com/solomonczyk/izborator/internal/semantic"

type semanticAggregate struct {
	total      int64
	valid      int64
	missingSum int64
}

// RecordSemanticValidation captures semantic validation results for later metrics.
func (s *Service) RecordSemanticValidation(result semantic.SemanticValidationResult) {
	if s == nil || s.logger == nil {
		return
	}

	domain := result.Domain
	if domain == "" {
		domain = "unknown"
	}
	missingCount := len(result.MissingSemantic)

	s.semanticMu.Lock()
	if s.semanticAgg == nil {
		s.semanticAgg = make(map[string]*semanticAggregate)
	}
	agg := s.semanticAgg[domain]
	if agg == nil {
		agg = &semanticAggregate{}
		s.semanticAgg[domain] = agg
	}

	agg.total++
	if result.Valid {
		agg.valid++
	}
	agg.missingSum += int64(missingCount)
	total := agg.total
	valid := agg.valid
	missingSum := agg.missingSum
	logEvery := s.semanticLogEvery
	s.semanticMu.Unlock()

	s.logger.Debug("scrapingstats: semantic validation result", map[string]interface{}{
		"domain":                 domain,
		"valid":                  result.Valid,
		"missing_semantic_count": len(result.MissingSemantic),
		"present_semantic_count": len(result.PresentSemantic),
	})

	if logEvery > 0 && total%logEvery == 0 {
		validRate := 0.0
		avgMissing := 0.0
		if total > 0 {
			validRate = float64(valid) / float64(total)
			avgMissing = float64(missingSum) / float64(total)
		}
		requiredCount := 2.0
		semanticCoverage := 1.0
		if requiredCount > 0 {
			semanticCoverage = 1.0 - (avgMissing / requiredCount)
		}
		normalizationSuccess := 1.0
		qualityScore := 0.5*validRate + 0.3*semanticCoverage + 0.2*normalizationSuccess
		failed := make([]string, 0, 2)
		if domain == "goods" || domain == "services" {
			thresholds := s.qualityGates.Goods
			if domain == "services" {
				thresholds = s.qualityGates.Services
			}
			if thresholds.ValidRateMin > 0 && validRate < thresholds.ValidRateMin {
				failed = append(failed, "valid_rate")
			}
			if thresholds.QualityScoreMin > 0 && qualityScore < thresholds.QualityScoreMin {
				failed = append(failed, "quality_score")
			}
		}

		s.logger.Info("scrapingstats: semantic validation snapshot", map[string]interface{}{
			"domain":                domain,
			"total":                 total,
			"valid_rate":            validRate,
			"avg_missing":           avgMissing,
			"required_count":        requiredCount,
			"semantic_coverage":     semanticCoverage,
			"normalization_success": normalizationSuccess,
			"quality_score":         qualityScore,
		})
		if len(failed) > 0 {
			s.logger.Warn("scrapingstats: quality gate failed", map[string]interface{}{
				"domain":                domain,
				"quality_gate_failed":   true,
				"failed":                failed,
				"valid_rate":            validRate,
				"quality_score":         qualityScore,
				"semantic_coverage":     semanticCoverage,
				"normalization_success": normalizationSuccess,
			})
		}
	}
}
