import { createRoutine } from 'redux-saga-routines';

// SystemIntakes routines
export const fetchSystemIntakes = createRoutine('FETCH_SYSTEM_INTAKES');

// SystemIntake routines
export const fetchSystemIntake = createRoutine('FETCH_SYSTEM_INTAKE');
export const saveSystemIntake = createRoutine('PUT_SYSTEM_INTAKE');
export const storeSystemIntake = createRoutine('STORE_SYSTEM_INTAKE');
export const submitSystemIntake = createRoutine('SUBMIT_SYSTEM_INTAKE');
export const removeSystemIntake = createRoutine('REMOVE_SYSTEM_INTAKE');

// SystemShorts routines
export const fetchSystemShorts = createRoutine('FETCH_SYSTEM_SHORTS');

// BusinessCase routines
export const fetchBusinessCase = createRoutine('FETCH_BUSINESS_CASE');
export const fetchBusinessCases = createRoutine('FETCH_BUSINESS_CASES');
