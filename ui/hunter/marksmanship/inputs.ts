import * as InputHelpers from '../../core/components/input_helpers';
import { RotationType, Spec } from '../../core/proto/common';
import i18n from '../../i18n/config';

// Configuration for spec-specific UI elements on the settings tab.
// These don't need to be in a separate file but it keeps things cleaner.

export const MMRotationConfig = {
	inputs: [
		InputHelpers.makeRotationEnumInput<Spec.SpecMarksmanshipHunter, RotationType>({
			fieldName: 'type',
			label: i18n.t('rotation_tab.common.rotation_type.label'),
			values: [
				{ name: i18n.t('rotation_tab.common.rotation_type.single_target'), value: RotationType.SingleTarget },
				{ name: i18n.t('rotation_tab.common.rotation_type.aoe'), value: RotationType.Aoe },
			],
		}),
	],
};
