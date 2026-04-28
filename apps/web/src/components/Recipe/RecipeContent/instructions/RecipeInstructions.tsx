'use client';
import React, { useEffect, useState } from 'react';
import Stack from '@mui/material/Stack';
import { models_Instruction } from '@4ks/api-fetch';
import { useRecipeContext } from '@/providers/recipe-context';
import RecipeDraggableInstructions from './RecipeDraggableInstructions';
import { SectionTitle } from '../../SectionTitle';
import AddIcon from '@mui/icons-material/Add';
import IconButton from '@mui/material/IconButton';
import {
  handleListAdd,
  handleListChange,
  handleListDelete,
  handleListDragEnd,
} from '../dnd-functions';

type RecipeInstructionsProps = {
  instructions: models_Instruction[];
};

export default function RecipeInstructions(props: RecipeInstructionsProps) {
  const [instructions, setInstructions] = useState(props.instructions);
  const [mounted, setMounted] = useState(false);
  const rtx = useRecipeContext();

  useEffect(() => {
    // Keep the first client render identical to SSR output. The DnD tree only
    // becomes stable after mount, so we swap to it in an effect.
    setMounted(true);
  }, []);

  useEffect(() => {
    // skip if undefined
    if (!rtx?.recipe.currentRevision?.instructions) {
      return;
    }
    // skip if no change
    if (rtx?.recipe.currentRevision?.instructions == instructions) {
      return;
    }
    setInstructions(rtx?.recipe.currentRevision?.instructions);
  }, [instructions, rtx?.recipe.currentRevision?.instructions]);

  function refreshInstructions(i: models_Instruction[]) {
    rtx?.setInstructions(i);
    setInstructions(i);
  }

  const onDragEnd = handleListDragEnd<models_Instruction>(
    instructions,
    refreshInstructions
  );

  const handleInstructionAdd = () =>
    handleListAdd<models_Instruction>(instructions, refreshInstructions);

  const handleInstructionDelete = (index: number) =>
    handleListDelete<models_Instruction>(
      index,
      instructions,
      refreshInstructions
    );

  const handleInstructionChange = handleListChange<models_Instruction>(
    instructions,
    refreshInstructions
  );

  const fallback = (
    <ol className="Instructions" style={{ paddingInlineStart: '24px' }}>
      {props.instructions.map((instruction, index) => (
        <li key={`instruction_${index}_${instruction.id}`} style={{ padding: '4px 0', fontSize: 20 }}>
          {instruction.text}
        </li>
      ))}
    </ol>
  );

  return (
    <Stack>
      <Stack direction="row" spacing={2}>
        <SectionTitle value={'Instructions'} />
        <IconButton aria-label="add" onClick={handleInstructionAdd}>
          <AddIcon />
        </IconButton>
      </Stack>

      {!mounted ? (
        fallback
      ) : (
        <RecipeDraggableInstructions
          data={instructions}
          onDragEnd={onDragEnd}
          onDelete={handleInstructionDelete}
          onChange={handleInstructionChange}
        />
      )}
    </Stack>
  );
}
