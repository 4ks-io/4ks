'use client';
import React, { useEffect, useState } from 'react';
import { models_Ingredient } from '@4ks/api-fetch';
import Stack from '@mui/material/Stack';
import DragHandleIcon from '@mui/icons-material/DragIndicator';
import DeleteIcon from '@mui/icons-material/Delete';
import TextField from '@mui/material/TextField';
import Checkbox from '@mui/material/Checkbox';
import IconButton from '@mui/material/IconButton';

interface RecipeIngredientProps {
  index: number;
  data: models_Ingredient;
  isDragging: boolean;
  handleIngredientDelete?: (index: number) => void;
  handleIngredientChange?: (index: number, data: models_Ingredient) => void;
}

export default function RecipeIngredient({
  data,
  index,
  isDragging,
  handleIngredientDelete = () => {},
  handleIngredientChange = () => {},
}: RecipeIngredientProps) {
  const [quantity, setQuantity] = useState(data.quantity || '');
  const [name, setName] = useState(data.name || '');
  const [active, setActive] = useState(false);
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    // MUI renders a different structure for multiline fields, so defer that
    // client-only enhancement until after hydration.
    setMounted(true);
  }, []);

  function handleQuantityChange(event: React.ChangeEvent<HTMLInputElement>) {
    setQuantity(`${event.target.value}`);
  }

  function handleNameChange(event: React.ChangeEvent<HTMLInputElement>) {
    setName(`${event.target.value}`);
  }

  function handleDelete() {
    handleIngredientDelete(index);
  }

  function handleFocus() {
    setActive(true);
  }

  function handleBlur() {
    setActive(false);
    handleIngredientChange(index, {
      id: data.id,
      type: data.type,
      name,
      quantity,
    } as models_Ingredient);
  }

  const inputProps = { disableUnderline: true, readOnly: false };
  const htmlInputProps = {
    // Some browser extensions mutate native input attributes before React
    // hydrates. Suppress those attribute-only warnings for these editor fields.
    suppressHydrationWarning: true,
  };

  return (
    <Stack style={isDragging ? { backgroundColor: '#E65100' } : {}}>
      <Stack direction="row">
        {active ? (
          <DragHandleIcon sx={{ fontSize: 20, marginTop: 1, marginLeft: 1 }} />
        ) : (
          <Checkbox size="small" />
        )}
        <TextField
          fullWidth
          size="small"
          variant="standard"
          placeholder="-"
          value={quantity || data.quantity}
          onFocus={handleFocus}
          onBlur={handleBlur}
          onChange={handleQuantityChange}
          multiline={mounted}
          InputProps={inputProps}
          sx={{ paddingLeft: 1, width: '96px', paddingTop: '0.5em' }}
          inputProps={{ ...htmlInputProps, style: { fontSize: 20 } }}
        />
        <TextField
          fullWidth
          size="small"
          variant="standard"
          placeholder="-"
          value={name || data.name}
          onFocus={handleFocus}
          onBlur={handleBlur}
          onChange={handleNameChange}
          multiline={mounted}
          InputProps={inputProps}
          sx={{ paddingLeft: 1, paddingTop: '8px' }}
          inputProps={{ ...htmlInputProps, style: { fontSize: 20 } }}
        />
        {active && (
          <IconButton aria-label="delete">
            <DeleteIcon fontSize="small" onClick={handleDelete} />
          </IconButton>
        )}
      </Stack>
    </Stack>
  );
}
