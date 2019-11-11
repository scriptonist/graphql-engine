import React from 'react';
import Spinner from '../../../../Common/Spinner/Spinner';
import { connect } from 'react-redux';
import { isInputObjectType, isObjectType, isEnumType } from 'graphql';
import { deriveExistingType } from '../utils';
import Tooltip from './Tooltip';
import styles from './Styles.scss';
import { useIntrospectionSchema } from '../introspection';

const CloneType = ({ headers, toggleModal, handleClonedTypes }) => {
  const [nameSpace, setNamespace] = React.useState('_');
  const namespaceOnChange = e => setNamespace(e.target.value);

  const { schema, loading, error, introspect } = useIntrospectionSchema(
    headers
  );

  if (loading) return <Spinner />;

  if (error) {
    return (
      <div>
        Error introspecting schema.&nbsp;
        <a onClick={introspect}>Try again</a>
      </div>
    );
  }

  const cloneableTypes = Object.keys(schema._typeMap)
    .filter(t => {
      return (
        isInputObjectType(schema._typeMap[t]) ||
        isObjectType(schema._typeMap[t]) ||
        isEnumType(schema._typeMap[t])
      );
    })
    .sort((t1, t2) => t1.toLowerCase() > t2.toLowerCase());

  const onSelection = e => {
    const selectedType = e.target.value;
    if (selectedType === '') return;
    const newTypes = deriveExistingType(
      selectedType,
      schema._typeMap,
      nameSpace
    );
    handleClonedTypes(newTypes);
    toggleModal();
  };

  const dropdownTitle = nameSpace ? null : 'Please provide a namespace first.';

  console.log(nameSpace === '');

  const namespaceTooltipText =
    'Namespace is required so that the type you are cloning does not collide with the existing type in Hasura.';

  return (
    <div>
      <div
        className={`row ${styles.add_mar_bottom_mid} ${styles.display_flex}`}
      >
        <div className={'col-md-3'}>
          Namespace <Tooltip text={namespaceTooltipText} id="clone-namespace" />
        </div>
        <input
          type="text"
          value={nameSpace}
          onChange={namespaceOnChange}
          className={`form-control col-md-3 ${styles.inputWidth}`}
        />
      </div>
      <div
        className={`row ${styles.add_mar_bottom_mid} ${styles.display_flex}`}
      >
        <div className="col-md-3"> Type to clone</div>
        <select
          value=""
          className={`form-control col-md-3 ${styles.inputWidth}`}
          onChange={onSelection}
          disabled={nameSpace === ''}
          title={dropdownTitle}
        >
          <option value="">---select an existing type---</option>
          {cloneableTypes.map(t => {
            return (
              <option value={t} key={t}>
                {t}
              </option>
            );
          })}
        </select>
      </div>
    </div>
  );
};

const mapStateToprops = state => {
  return {
    headers: state.tables.dataHeaders,
  };
};

export default connect(mapStateToprops)(CloneType);
