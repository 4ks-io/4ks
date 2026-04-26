'use client';
import React, { useState, useEffect } from 'react';
import ContentCopyIcon from '@mui/icons-material/ContentCopy';
import Fab from '@mui/material/Fab';
import Tooltip from '@mui/material/Tooltip';
import { trpc } from '@/trpc/client';
import { useUser } from '@auth0/nextjs-auth0';
import Button from '@mui/material/Button';
import Dialog from '@mui/material/Dialog';
import DialogTitle from '@mui/material/DialogTitle';
import DialogContent from '@mui/material/DialogContent';
import DialogActions from '@mui/material/DialogActions';
import IconButton from '@mui/material/IconButton';
import CloseIcon from '@mui/icons-material/Close';
import Typography from '@mui/material/Typography';
import TextField from '@mui/material/TextField';
import Snackbar from '@mui/material/Snackbar';
import MuiAlert, { AlertProps } from '@mui/material/Alert';
import CircularProgress from '@mui/material/CircularProgress';
import { validateFetchURL } from '@/libs/fetch-url';

export default function FetchRecipeButton() {
  const { user } = useUser();

  const [open, setOpen] = useState(false);
  const [url, setUrl] = useState('');
  const [err, setErr] = useState(false);
  const [errMsg, setErrMsg] = useState('');
  const [processing, setProcessing] = useState(false);
  const [saveSucess, setSaveSucess] = useState(false);
  const [saveError, setSaveError] = useState(false);

  // trpc
  const formData = trpc.recipes.fetch.useMutation();

  useEffect(() => {
    if (!processing || !formData.isPending) {
      return;
    }
    const { isError } = formData;

    if (isError) {
      setSaveError(true);
      setProcessing(false);
      return;
    }

    setSaveSucess(true);
    setProcessing(false);
    return;
  }, [formData, processing]);

  function handleClickOpen() {
    setOpen(true);
  }

  function handleClose() {
    setOpen(false);
  }

  function handleFetch() {
    if (!url || url.length === 0) {
      return;
    }
    try {
      const normalizedURL = validateFetchURL(url);
      setErr(false);
      setErrMsg('');
      setProcessing(true);
      formData.mutate(normalizedURL);
    } catch (e: unknown) {
      setErr(true);
      setErrMsg(e instanceof Error ? e.message : 'Invalid URL');
      return;
    }
  }

  function handleAlertClose(
    event?: React.SyntheticEvent | Event,
    reason?: string
  ) {
    if (reason === 'clickaway') {
      return;
    }

    setSaveSucess(false);
    setSaveError(false);
  }

  function handleValueChange(e: React.ChangeEvent<HTMLInputElement>) {
    setSaveSucess(false);
    setUrl(e.target.value);
  }

  return (
    <>
      <Snackbar
        open={saveSucess}
        autoHideDuration={2000}
        onClose={handleAlertClose}
      >
        <Alert onClose={handleAlertClose} severity="success">
          Save successful!
        </Alert>
      </Snackbar>
      <Snackbar
        open={saveError}
        autoHideDuration={2000}
        onClose={handleAlertClose}
      >
        <Alert onClose={handleAlertClose} severity="error">
          Failed to save.
        </Alert>
      </Snackbar>
      <Dialog
        onClose={handleClose}
        aria-labelledby="customized-dialog-title"
        open={open}
      >
        <DialogTitle sx={{ m: 0, p: 2 }} id="customized-dialog-title">
          Copy Recipe
        </DialogTitle>
        <IconButton
          aria-label="close"
          onClick={handleClose}
          sx={{
            position: 'absolute',
            right: 8,
            top: 8,
            color: (theme) => theme.palette.grey[500],
          }}
        >
          <CloseIcon />
        </IconButton>
        <DialogContent dividers>
          <Typography gutterBottom>
            {!user
              ? 'Please login to copy recipes.'
              : "Provide recipe URL and we'll do our best to fetch a copy."}
          </Typography>
          {user && (
            <TextField
              autoFocus
              disabled={processing || saveSucess}
              value={url}
              onChange={handleValueChange}
              id="outlined-basic"
              label="Recipe URL"
              variant="outlined"
              sx={{ width: '100%' }}
              error={err}
              helperText={errMsg}
            />
          )}
        </DialogContent>
        <DialogActions>
          {processing || saveSucess ? (
            <CircularProgress />
          ) : (
            <Button autoFocus onClick={handleFetch} disabled={!user}>
              OK
            </Button>
          )}
        </DialogActions>
      </Dialog>
      <Tooltip title="Copy" placement="left">
        <Fab
          color="primary"
          aria-label="add"
          sx={{
            margin: 0,
            top: 'auto',
            right: 20,
            bottom: 80,
            left: 'auto',
            position: 'fixed',
          }}
          onClick={handleClickOpen}
        >
          <ContentCopyIcon />
        </Fab>
      </Tooltip>
    </>
  );
}

const Alert = React.forwardRef<HTMLDivElement, AlertProps>(function Alert(
  props,
  ref
) {
  return <MuiAlert elevation={6} ref={ref} variant="filled" {...props} />;
});
