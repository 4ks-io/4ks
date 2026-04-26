'use client';
import React, { useState, useEffect } from 'react';
import { trpc } from '@/trpc/client';
import {
  normalizeForURL,
  navigateToLogin,
} from '@/libs/navigation';
import { useRouter } from 'next/navigation';
import { useRecipeContext } from '@/providers/recipe-context';
import Snackbar from '@mui/material/Snackbar';
import Fab from '@mui/material/Fab';
import SaveIcon from '@mui/icons-material/Save';
import SaveAsIcon from '@mui/icons-material/SaveAs';
import CloseIcon from '@mui/icons-material/Close';
import CallSplitIcon from '@mui/icons-material/CallSplit';
import LockPersonIcon from '@mui/icons-material/LockPerson';
import MuiAlert, { AlertProps } from '@mui/material/Alert';
import Tooltip from '@mui/material/Tooltip';
import {
  models_Recipe,
  dtos_UpdateRecipe,
  dtos_CreateRecipe,
  models_User,
} from '@4ks/api-fetch';

interface RecipeEditingControlsProps {
  user: models_User | undefined;
  recipe: models_Recipe;
  create?: boolean;
}

export default function RecipeEditingControls({
  user,
  recipe,
}: RecipeEditingControlsProps) {
  const rtx = useRecipeContext();
  const router = useRouter();

  const forkData = trpc.recipes.fork.useMutation();
  const createData = trpc.recipes.create.useMutation();
  const updateData = trpc.recipes.update.useMutation();
  const [updateSubmit, setUpdateSubmit] = useState(false);
  const [createSubmit, setCreateSubmit] = useState(false);
  const [saveError, setSaveError] = useState(false);
  const [saveDisabled, setSaveDisabled] = useState(true);

  const isAuthenticated = !!user?.id;
  const isContributor = isRecipeContributor();
  const isNew = recipe?.id == '0';

  useEffect(() => {
    if (rtx.recipe.currentRevision?.name == '') {
      setSaveDisabled(false);
      return;
    }
    setSaveDisabled(!rtx.editInProgress);
  }, [rtx.editInProgress, rtx]);

  // handle createData mutation effects
  useEffect(() => {
    const { isPending, isError } = createData;

    if (!createSubmit || isPending) {
      return;
    }

    // prevent infinite loop
    setCreateSubmit(false);

    if (isError) {
      setSaveError(true);
      return;
    }

    setSaveError(false);
    router.push(
      `/recipe/${createData.data?.id}-${normalizeForURL(
        createData.data?.currentRevision?.name
      )}`
    );

    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [createData, createSubmit]);

  // handle updateData mutation effects
  useEffect(() => {
    const { isPending, isError } = updateData;

    if (!updateSubmit || isPending) {
      return;
    }

    // prevent infinite loop
    setUpdateSubmit(false);
    rtx.setActionInProgress(false);

    if (isError) {
      setSaveError(true);
      return;
    }

    setSaveError(false);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [updateData, updateSubmit]);

  useEffect(() => {
    if (forkData.isSuccess) {
      router.push(
        `/recipe/${forkData.data?.id}-${normalizeForURL(
          forkData.data?.currentRevision?.name
        )}`
      );
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [forkData]);

  function isRecipeContributor() {
    if (!user || !user.id) return false;
    if (!recipe?.contributors) return false;

    const isContributor = recipe.contributors?.some((c) => {
      return user.id == c.id;
    });
    if (isContributor) return true;

    return false;
  }

  // contributors can edit
  async function saveRecipeChanges() {
    // double check, button shouldn't be visible if not a contributor
    if (isContributor) {
      setUpdateSubmit(true);
      rtx.setEditInProgress(false);
      rtx.setActionInProgress(true);
      updateData.mutate({
        recipeID: `${rtx?.recipeId}`,
        payload: rtx?.recipe.currentRevision as dtos_UpdateRecipe as any,
      });
    }
  }

  async function createRecipe() {
    setCreateSubmit(true);
    rtx.setEditInProgress(false);
    rtx.setActionInProgress(true);
    createData.mutate(rtx?.recipe.currentRevision as dtos_CreateRecipe as any);
  }

  function discardRecipeChanges() {
    rtx?.resetRecipe();
  }

  function handleClose(event?: React.SyntheticEvent | Event, reason?: string) {
    if (reason === 'clickaway') {
      return;
    }

    setSaveError(false);
  }

  function forkThisRecipe() {
    forkData.mutate(rtx?.recipeId);
  }

  function FabOptions() {
    const discard = (
      <Tooltip title="Discard changes">
        <Fab
          size="small"
          color="secondary"
          aria-label="discard"
          sx={{
            margin: 0,
            top: 'auto',
            right: 80,
            bottom: 20,
            left: 'auto',
            position: 'fixed',
          }}
          onClick={discardRecipeChanges}
        >
          <CloseIcon />
        </Fab>
      </Tooltip>
    );

    const save = (
      <Tooltip title="Save">
        <Fab
          disabled={saveDisabled}
          color="primary"
          aria-label="save"
          sx={{
            margin: 0,
            top: 'auto',
            right: 20,
            bottom: 20,
            left: 'auto',
            position: 'fixed',
          }}
          onClick={createRecipe}
        >
          <SaveIcon />
        </Fab>
      </Tooltip>
    );

    const edit = (
      <Tooltip title="Edit">
        <Fab
          disabled={saveDisabled}
          color="primary"
          aria-label="edit"
          sx={{
            margin: 0,
            top: 'auto',
            right: 20,
            bottom: 20,
            left: 'auto',
            position: 'fixed',
          }}
          onClick={saveRecipeChanges}
        >
          <SaveAsIcon />
        </Fab>
      </Tooltip>
    );

    const fork = (
      <Tooltip title="Fork to make your own">
        <Fab
          color="primary"
          aria-label="fork"
          sx={{
            margin: 0,
            top: 'auto',
            right: 20,
            bottom: 20,
            left: 'auto',
            position: 'fixed',
          }}
          onClick={forkThisRecipe}
        >
          <CallSplitIcon />
        </Fab>
      </Tooltip>
    );

    const login = (
      <Tooltip title="Login to save">
        <Fab
          color="primary"
          aria-label="fork"
          sx={{
            margin: 0,
            top: 'auto',
            right: 20,
            bottom: 20,
            left: 'auto',
            position: 'fixed',
          }}
          onClick={() => navigateToLogin()}
        >
          <LockPersonIcon />
        </Fab>
      </Tooltip>
    );

    return (
      <>
        {!isNew && rtx.editInProgress && discard}
        {!isAuthenticated && login}
        {isAuthenticated && isNew && save}
        {isAuthenticated && isContributor && edit}
        {isAuthenticated && !isNew && !isContributor && fork}
      </>
    );
  }

  return (
    <>
      <Snackbar open={saveError} autoHideDuration={2000} onClose={handleClose}>
        <Alert
          onClose={handleClose}
          severity="error"
          // sx={{ width: '100%' }}
        >
          Failed to save.
        </Alert>
      </Snackbar>

      {FabOptions()}
    </>
  );
}

const Alert = React.forwardRef<HTMLDivElement, AlertProps>(function Alert(
  props,
  ref
) {
  return <MuiAlert elevation={6} ref={ref} variant="filled" {...props} />;
});
