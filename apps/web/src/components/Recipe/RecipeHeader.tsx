'use client';
import React, { useEffect, useState } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { trpc } from '@/trpc/client';
import {
  normalizeForURL,
  navigateToLogin,
} from '@/libs/navigation';
import { useRecipeContext } from '@/providers/recipe-context';
import RecipeMediaBanner from './RecipeMediaBanner/RecipeMediaBanner';
import { models_Recipe, models_User } from '@4ks/api-fetch';
import Button from '@mui/material/Button';
import StarOutlineIcon from '@mui/icons-material/StarOutline';
import CallSplitIcon from '@mui/icons-material/CallSplit';
import ShareIcon from '@mui/icons-material/Share';
import EditIcon from '@mui/icons-material/Edit';
import TextField from '@mui/material/TextField';
import Badge from '@mui/material/Badge';
import Snackbar from '@mui/material/Snackbar';
import MuiAlert, { AlertProps } from '@mui/material/Alert';
import Toolbar from '@mui/material/Toolbar';
import Tooltip from '@mui/material/Tooltip';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import Stack from '@mui/material/Stack';
// import Avatar from '@mui/material/Avatar';

type RecipeHeaderProps = {
  user: models_User | undefined;
  recipe: models_Recipe;
};

// tr@ck: better SSR with recipe
export function RecipeHeader({ user, recipe }: RecipeHeaderProps) {
  const rtx = useRecipeContext();
  const router = useRouter();
  const forkData = trpc.recipes.fork.useMutation();
  const starData = trpc.recipes.star.useMutation();

  const isAuthenticated = !!user?.id;
  const isNew = recipe?.id == '0';

  // tr@ck: i18n
  const GENERIC_TITLE = `Fancy Recipe Name`;

  // const [isLoading, setIsLoading] = useState(true);
  // const [isNew, setIsNew] = useState(false);
  const [title, setTitle] = useState(
    isNew ? GENERIC_TITLE : recipe.currentRevision?.name || ''
  );
  const [chefName, setChefName] = useState(
    recipe?.author?.username || user?.username || 'chef'
  );
  const [forkActionMutex, setForkActionMutex] = useState(false);
  const [starActionMutex, setStarActionMutex] = useState(false);
  const [titleFocus, setTitleFocus] = useState(false);
  const [shareSucess, setShareSucess] = useState(false);
  const userIsChef =
    recipe?.author?.username == user?.username || !recipe?.author?.username;

  function handleTitleFocus() {
    setTitleFocus(true);
    if (title == GENERIC_TITLE) {
      setTitle('');
    }
  }

  function handleTitleChange(event: React.ChangeEvent<HTMLInputElement>) {
    setTitle(event.target.value);
  }

  useEffect(() => {
    if (rtx?.recipe?.currentRevision?.name != '') {
      setTitle(`${rtx?.recipe?.currentRevision?.name}`);
    }
  }, [rtx?.recipe?.currentRevision]);

  useEffect(() => {
    if (!forkActionMutex || forkData.isPending) {
      return;
    }
    if (forkData.isSuccess) {
      router.push(`/recipe/${forkData.data?.id}-${normalizeForURL(title)}`);
    }

    setForkActionMutex(false);

    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [forkData]);

  useEffect(() => {
    if (!starActionMutex || starData.isPending) {
      return;
    }

    setStarActionMutex(false);
    rtx.setActionInProgress(false);

    if (starData.isSuccess) {
      // tr@ck: simply increment/decrease the stars
      rtx.resetRecipe();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [starData]);

  function forkThisRecipe() {
    if (!isAuthenticated) {
      navigateToLogin();
      return;
    }

    setForkActionMutex(true);
    rtx.setActionInProgress(true);
    forkData.mutate(rtx?.recipeId);
  }

  function starThisRecipe() {
    if (!isAuthenticated) {
      navigateToLogin();
      return;
    }
    setStarActionMutex(true);
    rtx.setActionInProgress(true);
    starData.mutate(rtx?.recipeId);
  }

  async function shareThisRecipe() {
    const shareDetails = {
      url: window.location.href,
      title,
      text: title,
    };
    if (navigator.share) {
      try {
        await navigator.share(shareDetails);
      } catch (err) {
        console.error(err);
      }
    } else {
      // browser
      navigator.clipboard.writeText(window.location.href);
      setShareSucess(true);
    }
  }

  function handleClose(event?: React.SyntheticEvent | Event, reason?: string) {
    if (reason === 'clickaway') {
      return;
    }

    setShareSucess(false);
  }

  function handleTitleBlur() {
    if (['', undefined].includes(title)) {
      setTitle(GENERIC_TITLE);
      setTitleFocus(false);
      return;
    }
    rtx.setEditInProgress(true);
    rtx?.setTitle(title);
    setTitleFocus(false);
  }

  return (
    <>
      {/* media banner */}
      <RecipeMediaBanner isNew={isNew} />

      {/* toolbar */}
      <Toolbar
        variant="dense"
        sx={{ backgroundColor: '#fff', paddingBottom: 2 }}
      >
        <Box sx={{ flexGrow: 1 }} />
        <Badge color="primary" badgeContent={rtx?.recipe?.metadata?.forks}>
          <Button
            variant="outlined"
            startIcon={<CallSplitIcon />}
            onClick={forkThisRecipe}
            disabled={isNew}
            sx={{ marginLeft: 1 }}
          >
            Fork
          </Button>
        </Badge>
        <Badge color="primary" badgeContent={rtx?.recipe?.metadata?.stars}>
          <Button
            variant="outlined"
            startIcon={<StarOutlineIcon />}
            onClick={starThisRecipe}
            disabled={isNew}
            sx={{ marginLeft: 1 }}
          >
            Star
          </Button>
        </Badge>
        <Button
          variant="outlined"
          startIcon={<ShareIcon />}
          onClick={shareThisRecipe}
          disabled={isNew}
          sx={{ marginLeft: 1 }}
        >
          Share
        </Button>
      </Toolbar>

      {/* chef name*/}
      {/* <Toolbar
        variant="dense"
        sx={{ backgroundColor: '#fff', paddingBottom: 2 }}
      >
        <Link
          href={`/${chefName}`}
          style={{ textDecoration: 'none', color: '#000' }}
        >
          <Avatar
            sx={{ width: 28, height: 28, position: 'relative', left: '-12px' }}
          >
            {chefName.substring(0, 1)}
          </Avatar>
        </Link> */}
      <Link
        href={`/${chefName}`}
        prefetch={false}
        style={{ textDecoration: 'none', color: '#000' }}
      >
        {userIsChef ? (
          <Tooltip placement="right-start" title="me">
            <Badge
              color="secondary"
              variant="dot"
              anchorOrigin={{
                vertical: 'top',
                horizontal: 'left',
              }}
            >
              <Typography variant="h6" gutterBottom>
                @{chefName}
              </Typography>
            </Badge>
          </Tooltip>
        ) : (
          <Typography variant="h6" gutterBottom>
            @{chefName}
          </Typography>
        )}
      </Link>
      {/* </Toolbar> */}

      <Stack sx={{ paddingBottom: 4 }}>
        <Tooltip placement="left-start" title={<EditIcon fontSize="small" />}>
          <TextField
            variant="standard"
            onFocus={handleTitleFocus}
            onChange={handleTitleChange}
            onBlur={handleTitleBlur}
            value={title}
            // variant="standard"
            size="small"
            sx={{ width: '100%' }}
            inputProps={{
              // Browser extensions can inject attributes onto inputs before
              // hydration. Suppress that attribute-only mismatch here without
              // masking actual tree-shape hydration bugs.
              suppressHydrationWarning: true,
              style: { fontSize: 28 },
            }} // font size of input text
            InputProps={{
              disableUnderline: true,
              // startAdornment: titleFocus ? (
              //   <InputAdornment position="end">
              //     <EditIcon />
              //   </InputAdornment>
              // ) : (
              //   <></>
              // ),
            }}
          />
        </Tooltip>
      </Stack>

      <Snackbar
        open={shareSucess}
        autoHideDuration={2000}
        onClose={handleClose}
      >
        <Alert onClose={handleClose} severity="success">
          Copied to clipboard!
        </Alert>
      </Snackbar>
    </>
  );
}

const Alert = React.forwardRef<HTMLDivElement, AlertProps>(function Alert(
  props,
  ref
) {
  return <MuiAlert elevation={6} ref={ref} variant="filled" {...props} />;
});
