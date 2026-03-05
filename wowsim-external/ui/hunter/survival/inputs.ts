import * as InputHelpers from '../../core/components/input_helpers.js';
import { RotationType, Spec } from '../../core/proto/common.js';
import i18n from '../../i18n/config.js';

// Configuration for spec-specific UI elements on the settings tab.
// These don't need to be in a separate file but it keeps things cleaner.

// export const SniperTrainingUptime = InputHelpers.makeSpecOptionsNumberInput<Spec.SpecSurvivalHunter>({
// 	fieldName: 'sniperTrainingUptime',
// 	label: 'ST Uptime (%)',
// 	labelTooltip: 'Uptime for the Sniper Training talent, as a percent of the fight duration.',
// 	percent: true,
// 	showWhen: (player: Player<Spec.SpecSurvivalHunter>) => player.getTalents().sniperTraining > 0,
// 	changeEmitter: (player: Player<Spec.SpecSurvivalHunter>) => TypedEvent.onAny([player.specOptionsChangeEmitter, player.talentsChangeEmitter]),
// });

export const SVRotationConfig = {
	inputs: [
		InputHelpers.makeRotationEnumInput<Spec.SpecSurvivalHunter, RotationType>({
			fieldName: 'type',
			label: i18n.t('rotation_tab.common.rotation_type.label'),
			values: [
				{ name: i18n.t('rotation_tab.common.rotation_type.single_target'), value: RotationType.SingleTarget },
				{ name: i18n.t('rotation_tab.common.rotation_type.aoe'), value: RotationType.Aoe },
			],
		}),
	],
};
