'use client';

import * as React from 'react';
import { trpc } from '@/trpc/client';
import type { dtos_KitchenPassResponse } from '@4ks/api-fetch';
import {
  Alert,
  Box,
  Button,
  CircularProgress,
  Dialog,
  DialogActions,
  DialogContent,
  DialogContentText,
  DialogTitle,
  Paper,
  Snackbar,
  Stack,
  Typography,
} from '@mui/material';

type SettingsKitchenPassProps = {
  initialKitchenPass?: dtos_KitchenPassResponse;
};

type SnackbarState = {
  open: boolean;
  message: string;
  severity: 'success' | 'error';
};

function formatDate(value?: string) {
  if (!value) {
    return undefined;
  }

  return new Intl.DateTimeFormat('en', {
    dateStyle: 'medium',
    timeStyle: 'short',
  }).format(new Date(value));
}

export default function SettingsKitchenPass({
  initialKitchenPass,
}: SettingsKitchenPassProps) {
  const [isHydrated, setIsHydrated] = React.useState(false);
  const utils = trpc.useUtils();
  const kitchenPassQuery = trpc.users.getKitchenPass.useQuery(undefined, {
    initialData: initialKitchenPass,
  });
  const createKitchenPass = trpc.users.createKitchenPass.useMutation({
    onSuccess: async () => {
      await utils.users.getKitchenPass.invalidate();
      setNotice({
        open: true,
        message: 'Kitchen Pass updated.',
        severity: 'success',
      });
      setRotateOpen(false);
    },
    onError: () => {
      setNotice({
        open: true,
        message: 'Unable to update Kitchen Pass right now.',
        severity: 'error',
      });
    },
  });
  const deleteKitchenPass = trpc.users.deleteKitchenPass.useMutation({
    onSuccess: async () => {
      await utils.users.getKitchenPass.invalidate();
      setNotice({
        open: true,
        message: 'Kitchen Pass disabled.',
        severity: 'success',
      });
      setDisableOpen(false);
    },
    onError: () => {
      setNotice({
        open: true,
        message: 'Unable to disable Kitchen Pass right now.',
        severity: 'error',
      });
    },
  });

  const [rotateOpen, setRotateOpen] = React.useState(false);
  const [disableOpen, setDisableOpen] = React.useState(false);
  const [notice, setNotice] = React.useState<SnackbarState>({
    open: false,
    message: '',
    severity: 'success',
  });

  React.useEffect(() => {
    setIsHydrated(true);
  }, []);

  React.useEffect(() => {
    console.info('[kitchen-pass] settings component state', {
      hasInitialKitchenPass: !!initialKitchenPass,
      initialEnabled: !!initialKitchenPass?.enabled,
      queryStatus: kitchenPassQuery.status,
      fetchStatus: kitchenPassQuery.fetchStatus,
      isLoading: kitchenPassQuery.isLoading,
      isFetching: kitchenPassQuery.isFetching,
      isError: kitchenPassQuery.isError,
      hasData: !!kitchenPassQuery.data,
      enabled: !!kitchenPassQuery.data?.enabled,
      hasCopyText: !!kitchenPassQuery.data?.copyText,
      error: kitchenPassQuery.error?.message,
    });
  }, [
    initialKitchenPass,
    kitchenPassQuery.status,
    kitchenPassQuery.fetchStatus,
    kitchenPassQuery.isLoading,
    kitchenPassQuery.isFetching,
    kitchenPassQuery.isError,
    kitchenPassQuery.data,
    kitchenPassQuery.error,
  ]);

  const kitchenPass = kitchenPassQuery.data;
  const createdDate = isHydrated ? formatDate(kitchenPass?.createdDate) : undefined;
  const lastUsedDate = isHydrated ? formatDate(kitchenPass?.lastUsedDate) : undefined;
  const isBusy =
    kitchenPassQuery.isFetching ||
    createKitchenPass.isPending ||
    deleteKitchenPass.isPending;

  const handleCopy = async () => {
    const copyText = kitchenPass?.copyText;
    if (!copyText) {
      return;
    }

    try {
      await navigator.clipboard.writeText(copyText);
      setNotice({
        open: true,
        message: 'Instructions copied.',
        severity: 'success',
      });
    } catch {
      setNotice({
        open: true,
        message: 'Clipboard copy failed.',
        severity: 'error',
      });
    }
  };

  const handleGenerate = () => {
    createKitchenPass.mutate();
  };

  const handleCloseNotice = () => {
    setNotice((current) => ({ ...current, open: false }));
  };

  return (
    <>
      <Dialog open={rotateOpen} onClose={() => setRotateOpen(false)}>
        <DialogTitle>Rotate Kitchen Pass</DialogTitle>
        <DialogContent>
          <DialogContentText>
            Rotating creates a new AI Kitchen Pass and immediately disables the
            current one. You will need to paste the new instructions into your
            AI chat again.
          </DialogContentText>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setRotateOpen(false)}>Cancel</Button>
          <Button onClick={handleGenerate} autoFocus>
            Rotate Kitchen Pass
          </Button>
        </DialogActions>
      </Dialog>

      <Dialog open={disableOpen} onClose={() => setDisableOpen(false)}>
        <DialogTitle>Disable Kitchen Pass</DialogTitle>
        <DialogContent>
          <DialogContentText>
            Disabling removes your AI Kitchen Pass immediately. Any current AI
            session using it will stop working.
          </DialogContentText>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setDisableOpen(false)}>Cancel</Button>
          <Button color="error" onClick={() => deleteKitchenPass.mutate()} autoFocus>
            Disable Kitchen Pass
          </Button>
        </DialogActions>
      </Dialog>

      <Snackbar
        open={notice.open}
        autoHideDuration={2500}
        onClose={handleCloseNotice}
      >
        <Alert
          onClose={handleCloseNotice}
          severity={notice.severity}
          variant="filled"
        >
          {notice.message}
        </Alert>
      </Snackbar>

      <Paper
        variant="outlined"
        sx={{
          p: 3,
          borderRadius: 3,
        }}
      >
        <Stack spacing={2.5}>
          <Box>
            <Typography variant="h5" component="h3">
              Sous-Chef Kitchen Pass
            </Typography>
            <Typography variant="body1" color="text.secondary" sx={{ mt: 1 }}>
              Let your favorite AI sous-chef read and update your 4ks recipes.
            </Typography>
          </Box>

          {kitchenPassQuery.isLoading ? (
            <Box sx={{ display: 'flex', justifyContent: 'center', py: 2 }}>
              <CircularProgress size={28} />
            </Box>
          ) : kitchenPass?.enabled ? (
            <Stack spacing={2}>
              <Typography variant="subtitle1">
                Copy this into your AI chat:
              </Typography>
              <Paper
                variant="outlined"
                sx={{
                  p: 2,
                  whiteSpace: 'pre-wrap',
                  overflowWrap: 'anywhere',
                  wordBreak: 'break-word',
                  fontFamily: 'Monaco, Menlo, Consolas, monospace',
                  fontSize: 14,
                  lineHeight: 1.5,
                  bgcolor: 'grey.50',
                }}
              >
                {kitchenPass.copyText}
              </Paper>

              <Stack direction={{ xs: 'column', sm: 'row' }} spacing={1.5}>
                <Button
                  variant="contained"
                  onClick={handleCopy}
                  disabled={isBusy || !kitchenPass.copyText}
                >
                  Copy instructions
                </Button>
                <Button
                  variant="outlined"
                  onClick={() => setRotateOpen(true)}
                  disabled={isBusy}
                >
                  Rotate Kitchen Pass
                </Button>
                <Button
                  variant="text"
                  color="error"
                  onClick={() => setDisableOpen(true)}
                  disabled={isBusy}
                >
                  Disable Kitchen Pass
                </Button>
              </Stack>

              <Stack spacing={0.5}>
                <Typography variant="body2" color="text.secondary">
                  Status: Enabled
                </Typography>
                {createdDate && (
                  <Typography variant="body2" color="text.secondary">
                    Created: {createdDate}
                  </Typography>
                )}
                {lastUsedDate && (
                  <Typography variant="body2" color="text.secondary">
                    Last used: {lastUsedDate}
                    {kitchenPass.lastUsedAction
                      ? ` for ${kitchenPass.lastUsedAction}`
                      : ''}
                  </Typography>
                )}
              </Stack>

              <Alert severity="warning" variant="outlined">
                Anyone with this URL can read, create, update, and fork recipes
                on your behalf. Only share it with AI tools you trust.
              </Alert>
            </Stack>
          ) : (
            <Stack spacing={2}>
              <Typography variant="body1">
                No Kitchen Pass has been created yet.
              </Typography>
              <Typography variant="body2" color="text.secondary">
                After generating, copy the instructions and paste them into your
                AI chat.
              </Typography>
              <Box>
                <Button
                  variant="contained"
                  onClick={handleGenerate}
                  disabled={isBusy}
                >
                  Generate Kitchen Pass
                </Button>
              </Box>
            </Stack>
          )}
        </Stack>
      </Paper>
    </>
  );
}
