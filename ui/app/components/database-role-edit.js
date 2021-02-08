import Component from '@glimmer/component';
import { inject as service } from '@ember/service';
import { action } from '@ember/object';

const LIST_ROOT_ROUTE = 'vault.cluster.secrets.backend.list-root';
const SHOW_ROUTE = 'vault.cluster.secrets.backend.show';

export default class DatabaseRoleEdit extends Component {
  @service router;

  get tooltipMessage() {
    if (this.args.model.canUpdateDb === false) {
      return `You don't have permissions to update the connection "${
        this.args.model.database[0]
      }", so this role cannot be created.`;
    }
    if (!this.args.model.canCreateDynamic && this.args.model.type === 'dynamic') {
      return `You don't have permissions to create dynamic roles.`;
    }
    if (!this.args.model.canCreateStatic && this.args.model.type === 'static') {
      return `You don't have permissions to create static roles.`;
    }
    return null;
  }

  get databaseType() {
    if (this.args.model?.database) {
      // TODO: Calculate this
      return 'mongodb-database-plugin';
    }
    return null;
  }

  @action
  generateCreds(roleId) {
    this.router.transitionTo('vault.cluster.secrets.backend.credentials', roleId);
  }

  @action
  delete() {
    const secret = this.args.model;
    const backend = secret.backend;
    secret.destroyRecord().then(() => {
      this.router.transitionTo(LIST_ROOT_ROUTE, backend, { queryParams: { tab: 'role' } });
    });
  }

  @action
  handleCreateRole(evt) {
    evt.preventDefault();
    let roleSecret = this.args.model;
    let secretId = roleSecret.name;
    roleSecret.set('id', secretId);
    let path = roleSecret.type === 'static' ? 'static-roles' : 'roles';
    roleSecret.set('path', path);
    roleSecret.save().then(() => {
      this.router.transitionTo(SHOW_ROUTE, `role/${secretId}`);
    });
  }

  @action
  handleCreateEditRole(evt) {
    evt.preventDefault();
    const mode = this.args.mode;
    let roleSecret = this.args.model;
    let secretId = roleSecret.name;
    if (mode === 'create') {
      roleSecret.set('id', secretId);
      let path = roleSecret.type === 'static' ? 'static-roles' : 'roles';
      roleSecret.set('path', path);
    }
    roleSecret.save().then(() => {
      this.router.transitionTo(SHOW_ROUTE, `role/${secretId}`);
    });
  }
}
