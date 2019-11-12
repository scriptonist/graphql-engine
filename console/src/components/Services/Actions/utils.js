// type utils

export const findType = (types, typeName) => {
  return types.find(t => t.name === typeName);
};

// action utils

export const getActionName = action => {
  return action.action_name;
};

export const findAction = (actions, actionName) => {
  return actions.find(a => getActionName(a) === actionName);
};

export const getActionOutputType = action => {
  return action.action_defn.output_type;
};

export const getActionOutputFields = (action, types) => {
  const outputTypeName = getActionOutputType(action);

  const outputType = findType(types, outputTypeName);

  return outputType.fields;
};

export const getActionArguments = action => {
  return action.action_defn.arguments;
};

export const getAllActions = getState => {
  return getState().actions.common.actions;
};
