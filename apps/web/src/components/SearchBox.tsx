'use client';
import React from 'react';
import SearchIcon from '@mui/icons-material/Search';
import InputAdornment from '@mui/material/InputAdornment';
import TextField from '@mui/material/TextField';
import Chip from '@mui/material/Chip';
import { isMac, isMobileDevice } from '@/libs/navigation';
import Tooltip from '@mui/material/Tooltip';
import CloseIcon from '@mui/icons-material/Close';

type SearchBoxProps = {
  value: string;
  onClick: () => void;
  onClear: () => void;
};

export default function SearchBox({ value, onClick, onClear }: SearchBoxProps) {
  const showShortcut = !isMobileDevice();

  return (
    <TextField
      id="searchBox"
      autoComplete="off"
      placeholder="Search..."
      size="small"
      sx={{ width: 300 }}
      InputProps={{
        startAdornment: (
          <>
            {(!value || value == '') && (
              <InputAdornment position="start">
                <SearchIcon />
              </InputAdornment>
            )}
          </>
        ),
        endAdornment: (
          <InputAdornment position="end">
            {showShortcut && (
              <Chip
                label={isMac() ? '⌘+K' : 'Ctrl+K'}
                size="small"
                variant="outlined"
              />
            )}
            {value && value != '' && (
              <Tooltip title="Clear">
                <CloseIcon
                  fontSize="small"
                  sx={{ cursor: 'pointer' }}
                  onClick={onClear}
                />
              </Tooltip>
            )}
          </InputAdornment>
        ),
      }}
      value={value}
      onClick={onClick}
    />
  );
}
