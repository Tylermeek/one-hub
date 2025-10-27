import { combineReducers } from 'redux';

// reducer import
import customizationReducer from './customizationReducer';
import accountReducer from './accountReducer';
import siteInfoReducer from './siteInfoReducer';
import contextReducer from './contextReducer';

// ==============================|| COMBINE REDUCER ||============================== //

const reducer = combineReducers({
    customization: customizationReducer,
    account: accountReducer,
    siteInfo: siteInfoReducer,
    context: contextReducer
});

export default reducer;
