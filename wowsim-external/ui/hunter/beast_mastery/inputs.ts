import * as InputHelpers from '../../core/components/input_helpers';
import { Player } from '../../core/player';
import { RotationType, Spec } from '../../core/proto/common';
import { HunterStingType as StingType } from '../../core/proto/hunter';
import { TypedEvent } from '../../core/typed_event';
import i18n from '../../i18n/config';

// Configuration for spec-specific UI elements on the settings tab.
// These don't need to be in a separate file but it keeps things cleaner.

export const BMRotationConfig = {
	inputs: [
		InputHelpers.makeRotationEnumInput<Spec.SpecBeastMasteryHunter, RotationType>({
			fieldName: 'type',
			label: i18n.t('rotation_tab.common.rotation_type.label'),
			values: [
				{ name: i18n.t('rotation_tab.common.rotation_type.single_target'), value: RotationType.SingleTarget },
				{ name: i18n.t('rotation_tab.common.rotation_type.aoe'), value: RotationType.Aoe },
			],
		}),
		InputHelpers.makeRotationEnumInput<Spec.SpecBeastMasteryHunter, StingType>({
			fieldName: 'sting',
			label: i18n.t('rotation_tab.options.hunter.beast_mastery.sting.label'),
			labelTooltip: i18n.t('rotation_tab.options.hunter.beast_mastery.sting.tooltip'),
			values: [
				{ name: i18n.t('rotation_tab.options.hunter.beast_mastery.sting.values.none'), value: StingType.NoSting },
				{ name: i18n.t('rotation_tab.options.hunter.beast_mastery.sting.values.serpent_sting'), value: StingType.SerpentSting },
			],
			showWhen: (player: Player<Spec.SpecBeastMasteryHunter>) => player.getSimpleRotation().type == RotationType.SingleTarget,
		}),
		InputHelpers.makeRotationBooleanInput<Spec.SpecBeastMasteryHunter>({
			fieldName: 'trapWeave',
			label: i18n.t('rotation_tab.options.hunter.beast_mastery.trap_weave.label'),
			labelTooltip: i18n.t('rotation_tab.options.hunter.beast_mastery.trap_weave.tooltip'),
		}),
		InputHelpers.makeRotationBooleanInput<Spec.SpecBeastMasteryHunter>({
			fieldName: 'multiDotSerpentSting',
			label: i18n.t('rotation_tab.options.hunter.beast_mastery.multi_dot_serpent_sting.label'),
			labelTooltip: i18n.t('rotation_tab.options.hunter.beast_mastery.multi_dot_serpent_sting.tooltip'),
			changeEmitter: (player: Player<Spec.SpecBeastMasteryHunter>) => TypedEvent.onAny([player.rotationChangeEmitter, player.talentsChangeEmitter]),
		}),
	],
};
