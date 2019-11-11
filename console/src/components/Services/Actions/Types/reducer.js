import defaultState from './state';

const SET_TYPE_DEFINITION = 'Actions/Types/SET_TYPE_DEFINITION';
export const setTypeDefinition = (sdl, error = null) => ({
  type: SET_TYPE_DEFINITION,
  definition: { sdl, error },
});

const SET_FETCHING = 'Actions/Types/SET_FETCHING';
export const setFetching = () => ({ type: SET_FETCHING });
const UNSET_FETCHING = 'Actions/Types/UNSET_FETCHING';
export const unsetFetching = () => ({ type: UNSET_FETCHING });

const reducer = (state = defaultState, action) => {
  switch (action.type) {
    case SET_FETCHING:
      return {
        ...state,
        isFetching: true,
      };
    case UNSET_FETCHING:
      return {
        ...state,
        isFetching: false,
      };
    case SET_TYPE_DEFINITION:
      return {
        ...state,
        manage: {
          definition: action.definition,
        },
      };
    default:
      return state;
  }
};

export default reducer;
