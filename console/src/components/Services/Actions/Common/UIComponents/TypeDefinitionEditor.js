import React from 'react';
import { parse as sdlParse } from 'graphql/language/parser';
import styles from './Styles.scss';
import Tooltip from './Tooltip';
import CrossIcon from '../../../../Common/Icons/Cross';
import CopyIcon from '../../../../Common/Icons/Copy';
import SDLEditor from '../../../../Common/AceEditor/SDLEditor';
import Modal from '../../../../Common/Modal/Modal';
import CloneTypeModal from './CloneTypeModal';
import { getTypesSdl } from '../../../Types/sdlUtils';

const editorLabel = 'Type Definition';
const editorTooltip = 'Define your action as a GraphQL mutation using SDL';

let parseDebounceTimer = null;

const ActionDefinitionEditor = ({
  value,
  onChange,
  className,
  placeholder,
  error,
}) => {
  const [modalOpen, setModalState] = React.useState(false);
  const toggleModal = () => setModalState(!modalOpen);

  const onChangeWithError = v => {
    if (parseDebounceTimer) {
      clearTimeout(parseDebounceTimer);
    }
    parseDebounceTimer = setTimeout(() => {
      if (!v.trim()) return;
      let _e = null;
      try {
        sdlParse(v);
      } catch (e) {
        _e = e;
      }
      if (_e) {
        onChange(v, _e);
      }
    }, 1000);

    onChange(v);
  };

  const errorMessage =
    error && (error.message || 'This is not valid GraphQL SDL');

  let markers = [];
  if (error && error.locations) {
    markers = error.locations.map(l => ({
      line: l.line,
      column: l.column,
      type: 'error',
      message: errorMessage,
      className: styles.errorMarker,
    }));
  }

  const handleClonedTypes = types => {
    onChange(`${value}\n\n${getTypesSdl(types)}`);
  };

  return (
    <div className={`${className || ''}`}>
      <h2
        className={`${styles.subheading_text} ${styles.add_mar_bottom_small}`}
      >
        {editorLabel}
        <Tooltip
          id="action-name"
          text={editorTooltip}
          className={styles.add_mar_left_mid}
        />
      </h2>
      <div className={styles.sdlEditorContainer}>
        <div
          className={`${styles.display_flex} ${styles.add_mar_bottom_small}`}
        >
          {error && (
            <div className={`${styles.display_flex} ${styles.errorMessage}`}>
              <CrossIcon className={styles.add_mar_right_small} />
              <div>{errorMessage}</div>
            </div>
          )}
          <a
            className={`${styles.cloneTypeText} ${styles.cursorPointer}`}
            onClick={toggleModal}
          >
            <CopyIcon className={styles.add_mar_right_small} />
            Clone an existing type
          </a>
          <Modal
            show={modalOpen}
            title={'Clone an existing type'}
            onClose={toggleModal}
            customClass={styles.modal}
          >
            <CloneTypeModal
              handleClonedTypes={handleClonedTypes}
              toggleModal={toggleModal}
            />
          </Modal>
        </div>
        <SDLEditor
          name="sdl-editor"
          value={value}
          onChange={onChangeWithError}
          placeholder={placeholder}
          markers={markers}
          height="200px"
          width="600px"
        />
      </div>
    </div>
  );
};

export default ActionDefinitionEditor;
