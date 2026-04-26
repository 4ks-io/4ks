'use client';
import React, { useEffect, useContext, useReducer, useRef } from 'react';
import { initialState } from './search-context-init';
import { SearchContextState } from './search-context-types';
import {
  searchContextReducer,
  SearchContextAction,
} from './search-context-reducer';
import TypesenseInstantSearchAdapter from 'typesense-instantsearch-adapter';
import log from '@/libs/logger';
import { isMobileDevice } from '@/libs/navigation';

const SearchContext = React.createContext<SearchContextState>(initialState);

type SearchContextProviderProps = {
  children: React.ReactNode;
  typesenseApikey: string;
  typesenseUrl: string;
  typesensePath?: string | undefined;
};

export function SearchContextProvider({
  children,
  typesenseApikey,
  typesenseUrl,
  typesensePath,
}: SearchContextProviderProps) {
  const [state, dispatch] = useReducer(searchContextReducer, initialState);
  const isOpenRef = useRef(state.open);

  useEffect(() => {
    isOpenRef.current = state.open;
  }, [state.open]);

  useEffect(() => {
    log().Debug(new Error(), [{ k: 'SearchContextProvider', v: 'init' }]);
    function onKeyDown(event: KeyboardEvent) {
      if (isMobileDevice()) {
        return;
      }

      if (event.ctrlKey && event.key === 'k' && !isOpenRef.current) {
        handleOpen();
        event.preventDefault();
      }
    }

    window.addEventListener('keydown', onKeyDown);

    return () => {
      window.removeEventListener('keydown', onKeyDown);
    };
  }, []);

  function setResults(results: []) {
    dispatch({
      type: SearchContextAction.SET_RESULTS,
      payload: results,
    });
  }

  function setValue(value: string) {
    dispatch({
      type: SearchContextAction.SET_VALUE,
      payload: value,
    });
  }

  function clear() {
    dispatch({
      type: SearchContextAction.CLEAR,
      payload: null,
    });
  }

  function clearResults() {
    setResults([]);
  }

  function handleOpen() {
    dispatch({
      type: SearchContextAction.OPEN_DIALOG,
      payload: null,
    });
  }

  function handleClose() {
    dispatch({
      type: SearchContextAction.CLOSE_DIALOG,
      payload: null,
    });
  }

  function getTypeseneNode() {
    if (typesensePath && typesensePath != '') {
      return {
        host: typesenseUrl,
        path: typesensePath,
        port: 443,
        protocol: 'https',
      };
    }

    return {
      host: typesenseUrl,
      port: 443,
      protocol: 'https',
    };
  }

  useEffect(() => {
    const typesenseInstantsearchAdapter = new TypesenseInstantSearchAdapter({
      server: {
        apiKey: typesenseApikey,
        nodes: [getTypeseneNode()],
        // Cache search results from server. Defaults to 2 minutes. Set to 0 to disable caching.
        cacheSearchResultsForSeconds: 2 * 60,
      },
      additionalSearchParameters: {
        query_by: 'name,author,ingredients',
        per_page: 21,
      },
    });

    dispatch({
      type: SearchContextAction.INIT,
      payload: {
        client: typesenseInstantsearchAdapter.searchClient,
        open: false,
        handleClose,
        handleOpen,
        setResults,
        clearResults,
        clear,
        setValue,
      },
    });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  return (
    <SearchContext.Provider value={state}>{children}</SearchContext.Provider>
  );
}

export function useSearchContext() {
  return useContext(SearchContext);
}
