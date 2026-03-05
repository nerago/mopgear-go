import * as InputHelpers from '../../core/components/input_helpers.js';
import { Spec } from '../../core/proto/common.js';
import i18n from '../../i18n/config.js';

// Configuration for spec-specific UI elements on the settings tab.
// These don't need to be in a separate file but it keeps things cleaner.

export const OkfUptime = InputHelpers.makeSpecOptionsNumberInput<Spec.SpecBalanceDruid>({
	fieldName: 'okfUptime',
	label: i18n.t('settings_tab.other.okf_uptime.label'),
	labelTooltip: i18n.t('settings_tab.other.okf_uptime.tooltip'),
	percent: true,
});
