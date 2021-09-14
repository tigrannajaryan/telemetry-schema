package schema

import (
	"sort"

	"github.com/tigrannajaryan/telemetry-schema/schema/ast"
	"github.com/tigrannajaryan/telemetry-schema/schema/compiled"
	"github.com/tigrannajaryan/telemetry-schema/schema/types"
)

func Compile(schema *ast.Schema) *compiled.Schema {
	compiledSchema := &compiled.Schema{}

	compiledActionsForVersion := map[types.TelemetryVersion]*compiled.ActionsForVersion{}

	// Loop through and compile each version.
	for versionNum, versionDescr := range schema.Versions {
		actionsForVer, exists := compiledActionsForVersion[versionNum]
		if !exists {
			actionsForVer = &compiled.ActionsForVersion{}
			compiledActionsForVersion[versionNum] = actionsForVer
		}

		actionsForVer.Resource = compileResourceActions(versionDescr.All.Changes, versionDescr.Resources.Changes)
		actionsForVer.Metrics = compileMetricActions(versionDescr.All.Changes, versionDescr.Metrics.Changes)
		actionsForVer.Spans = compileSpanActions(versionDescr.All.Changes, versionDescr.Spans.Changes)
	}

	// Convert map by version to a slice.
	for versionNum, actions := range compiledActionsForVersion {
		actions.VersionNum = versionNum
		compiledSchema.Versions = append(compiledSchema.Versions, actions)
	}

	// Order the slice by version.
	sort.Sort(compiledSchema.Versions)

	return compiledSchema
}

func compileResourceActions(
	allActions []ast.AttributeTranslationAction,
	resourceActions []ast.AttributeTranslationAction,
) (result compiled.ResourceActions) {

	var compiledActionSeq []compiled.ResourceAction

	// First add actions in "all" section.
	for _, action := range allActions {
		if action.RenameAttributes != nil {
			compiledAction := compiled.ResourceAttributesRenameAction(*action.RenameAttributes)
			compiledActionSeq = append(compiledActionSeq, compiledAction)
		}
	}

	// Now compile resource actions and add one by one.
	for _, action := range resourceActions {
		if action.RenameAttributes != nil {
			compiledAction := compiled.ResourceAttributesRenameAction(*action.RenameAttributes)
			compiledActionSeq = append(compiledActionSeq, compiledAction)
		}
	}

	return compiledActionSeq
}

func compileMetricActions(
	allActions []ast.AttributeTranslationAction,
	metricActions []ast.MetricTranslationAction,
) (result compiled.MetricActions) {

	var compiledActionSeq []compiled.MetricAction

	// First add actions in "all" section.
	for _, action := range allActions {
		if action.RenameAttributes != nil {
			compiledAction := compiled.MetricLabelRenameAction{
				LabelMap: *action.RenameAttributes,
			}
			compiledActionSeq = append(compiledActionSeq, compiledAction)
			// Should apply to all metrics.
			result.OtherMetrics = append(result.OtherMetrics, compiledAction)
		}
	}

	// Now compile metric actions and add one by one.
	affectedMetrics := map[types.MetricName]bool{}
	for _, srcAction := range metricActions {
		var compiledAction compiled.MetricAction

		if srcAction.RenameMetrics != nil {
			compiledAction = compiled.MetricRenameAction(srcAction.RenameMetrics)
			for metricName := range srcAction.RenameMetrics {
				affectedMetrics[metricName] = true
			}
		} else if srcAction.RenameLabels != nil {
			compiledAction = compiled.MetricLabelRenameAction{
				ApplyOnlyToMetrics: metricNamesToMap(srcAction.RenameLabels.ApplyToMetrics),
				LabelMap:           srcAction.RenameLabels.LabelMap,
			}

			if len(srcAction.RenameLabels.ApplyToMetrics) == 0 {
				// Should apply to all metrics.
				result.OtherMetrics = append(result.OtherMetrics, compiledAction)
			} else {
				// Applies to specific metrics only.
				for _, metricName := range srcAction.RenameLabels.ApplyToMetrics {
					affectedMetrics[metricName] = true
				}
			}
		}

		if compiledAction != nil {
			compiledActionSeq = append(compiledActionSeq, compiledAction)
		}
	}

	result.ByName = map[types.MetricName][]compiled.MetricAction{}

	for metricName := range affectedMetrics {
		result.ByName[metricName] = compiledActionSeq
		// TODO: optimize compiledActionSeq by checking if metricName is in the
		// ApplyOnlyToMetrics map that limits the application of particular action
		// then ApplyOnlyToMetrics can be deleted since it has no effect. That will
		// speed up the action execution since we no longer need to lookup the metric
		// name in the limit map.
	}

	return result
}

func metricNamesToMap(metrics []types.MetricName) map[types.MetricName]bool {
	m := map[types.MetricName]bool{}
	for _, metric := range metrics {
		m[metric] = true
	}
	return m
}

func compileSpanActions(
	allActions []ast.AttributeTranslationAction,
	spanActions []ast.SpanTranslationAction,
) (result compiled.SpanActions) {

	var compiledActionSeq []compiled.SpanAction

	// First add actions in "all" section.
	for _, action := range allActions {
		if action.RenameAttributes != nil {
			compiledAction := compiled.SpanAttributeRenameAction{
				AttributesRenameAction: map[string]string(*action.RenameAttributes),
			}
			compiledActionSeq = append(compiledActionSeq, compiledAction)
			// Should apply to all metrics.
			result.ForAllSpans = append(result.ForAllSpans, compiledAction)
		}
	}

	// Now compile span actions and add one by one.
	for _, srcAction := range spanActions {
		var compiledAction compiled.SpanAction

		if srcAction.RenameAttributes != nil {
			compiledAction = compiled.SpanAttributeRenameAction{
				AttributesRenameAction: srcAction.RenameAttributes.AttributeMap,
			}

			result.ForAllSpans = append(result.ForAllSpans, compiledAction)
		}

		if compiledAction != nil {
			compiledActionSeq = append(compiledActionSeq, compiledAction)
		}
	}

	return result
}

func spanNamesToMap(spans []types.SpanName) map[types.SpanName]bool {
	m := map[types.SpanName]bool{}
	for _, span := range spans {
		m[span] = true
	}
	return m
}
