import i18n from '../../../../i18n/config';
import { IndividualSimUI } from '../../../individual_sim_ui';
import { Player } from '../../../player';
import { TypedEvent } from '../../../typed_event';
import { Component } from '../../component';
import { ListPicker } from '../../pickers/list_picker';

export class AplFloatingActionBar extends Component {
	constructor(parent: HTMLElement, simUI: IndividualSimUI<any>, listPicker: ListPicker<Player<any>, any>, itemName: string) {
		super(parent, 'apl-floating-action-bar-root');

		const newButton = this.rootElem.appendChild(
			<button className="btn btn-primary">
				<i className="fas fa-plus me-2" />
				{i18n.t('rotation_tab.apl.floatingActionBar.new', { itemName: itemName })}
			</button>,
		);

		newButton.addEventListener('click', () => {
			const newItem = listPicker.config.newItem();
			const newList = listPicker.config.getValue(listPicker.modObject).concat([newItem]);
			listPicker.config.setValue(TypedEvent.nextEventID(), listPicker.modObject, newList);
		});

		const resetButton = this.rootElem.appendChild(
			<button className="btn btn-sm btn-link btn-reset ms-auto">
				<i className="fas fa-times me-1"></i>
				{i18n.t('rotation_tab.apl.floatingActionBar.reset')}
			</button>,
		);

		resetButton.addEventListener('click', () => {
			simUI.applyEmptyAplRotation(TypedEvent.nextEventID());
		});

		new IntersectionObserver(
			([e]) => {
				e.target.classList.toggle('stuck');
			},
			{
				root: parent,
				rootMargin: '-100% 0px 0px 0px',
			},
		).observe(this.rootElem);
	}
}
