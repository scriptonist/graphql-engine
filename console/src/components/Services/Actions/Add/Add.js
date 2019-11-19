import React from 'react';
import styles from './Styles.scss';
import Helmet from 'react-helmet';
import WebhookEditor from '../Common/UIComponents/WebhookEditor';
import KindEditor from '../Common/UIComponents/KindEditor';
import ActionDefinitionEditor from '../Common/UIComponents/ActionDefinitionEditor';
import TypeDefinitionEditor from '../Common/UIComponents/TypeDefinitionEditor';
import Button from '../../../Common/Button';
import {
  setActionWebhook,
  setActionKind,
  setDefaults,
  setActionDefinition,
  setTypeDefinition,
} from './reducer';
import { createAction } from '../ServerIO';

const AddAction = ({
  webhook,
  dispatch,
  kind,
  actionDefinition,
  typeDefinition,
  isFetching,
}) => {
  React.useEffect(() => {
    dispatch(setDefaults());
  }, []);

  const webhookOnChange = e => dispatch(setActionWebhook(e.target.value));
  const kindOnChange = k => dispatch(setActionKind(k));

  const {
    sdl: typesDefinitionSdl,
    error: typesDefinitionError,
    timer: typedefParseTimer,
  } = typeDefinition;

  const {
    sdl: actionDefinitionSdl,
    error: actionDefinitionError,
    timer: actionParseTimer,
  } = actionDefinition;

  const onSubmit = () => {
    dispatch(createAction());
  };

  const actionDefinitionOnChange = (value, error, timer, ast) => {
    dispatch(setActionDefinition(value, error, timer, ast));
  };

  const typeDefinitionOnChange = (value, error, timer, ast) => {
    dispatch(setTypeDefinition(value, error, timer, ast));
  };

  const allowSave =
    !isFetching &&
    !typesDefinitionError &&
    !actionDefinitionError &&
    !actionParseTimer &&
    !typedefParseTimer;

  return (
    <div>
      <Helmet title={'Add Action - Actions | Hasura'} />
      <div className={styles.heading_text}>Add a new action</div>
      <WebhookEditor
        value={webhook}
        onChange={webhookOnChange}
        placeholder="action webhook"
        className={styles.add_mar_bottom_mid}
        service="create-action"
      />
      <hr />
      <KindEditor
        value={kind}
        onChange={kindOnChange}
        className={styles.add_mar_bottom_mid}
      />
      <hr />
      <ActionDefinitionEditor
        value={actionDefinitionSdl}
        error={actionDefinitionError}
        onChange={actionDefinitionOnChange}
        timer={actionParseTimer}
        placeholder={''}
      />
      <hr />
      <TypeDefinitionEditor
        value={typesDefinitionSdl}
        error={typesDefinitionError}
        timer={typedefParseTimer}
        onChange={typeDefinitionOnChange}
        placeholder={''}
      />
      <hr />
      <Button
        color="yellow"
        size="sm"
        type="submit"
        disabled={!allowSave}
        onClick={onSubmit}
      >
        Create
      </Button>
    </div>
  );
};

export default AddAction;
