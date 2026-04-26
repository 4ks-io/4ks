import React, { useEffect, useState } from 'react';
import { models_Instruction } from '@4ks/api-fetch';
import Stack from '@mui/material/Stack';
import DragHandleIcon from '@mui/icons-material/DragIndicator';
import DeleteIcon from '@mui/icons-material/Delete';
import TextField from '@mui/material/TextField';
import IconButton from '@mui/material/IconButton';

interface RecipeInstructionProps {
  index: number;
  data: models_Instruction;
  isDragging: boolean;
  handleInstructionDelete?: (index: number) => void;
  handleInstructionChange?: (index: number, data: models_Instruction) => void;
}

export default function RecipeInstruction({
  data,
  index,
  isDragging,
  handleInstructionDelete = () => {},
  handleInstructionChange = () => {},
}: RecipeInstructionProps) {
  const [active, setActive] = useState(false);
  const [text, setText] = useState(data.text || '');
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    // MUI renders a different structure for multiline fields, so defer that
    // client-only enhancement until after hydration.
    setMounted(true);
  }, []);

  function handleTextChange(event: React.ChangeEvent<HTMLInputElement>) {
    setText(`${event.target.value}`);
  }

  function handleDelete() {
    handleInstructionDelete(index);
  }

  function handleFocus() {
    setActive(true);
  }

  function handleBlur() {
    setActive(false);
    handleInstructionChange(index, {
      id: data.id,
      type: data.type,
      name: '',
      text,
    } as models_Instruction);
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
        {!active && (
          <TextField
            variant="standard"
            value={index + 1}
            InputProps={{ disableUnderline: true, readOnly: true }}
            sx={{ width: 20 }}
            inputProps={{ ...htmlInputProps, style: { fontSize: 20 } }}
          />
        )}
        {active && (
          <DragHandleIcon
            sx={{ fontSize: 20, marginTop: 1, marginLeft: '2px' }}
          />
        )}
        <TextField
          fullWidth
          size="small"
          variant="standard"
          placeholder="-"
          value={text}
          onFocus={handleFocus}
          onBlur={handleBlur}
          onChange={handleTextChange}
          multiline={mounted}
          InputProps={inputProps}
          sx={{ paddingLeft: 1, paddingTop: '0.5em' }}
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
