'use client';
import React, { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { useUser } from '@auth0/nextjs-auth0';
import { trpc } from '@/trpc/client';
import { z } from 'zod';
import Button from '@mui/material/Button';
import TextField from '@mui/material/TextField';
import Typography from '@mui/material/Typography';
import MuiAlert, { AlertProps } from '@mui/material/Alert';
import CircularProgress from '@mui/material/CircularProgress';
import Container from '@mui/material/Container';
import Snackbar from '@mui/material/Snackbar';
import Box from '@mui/material/Box';

type InputValidation = {
  Err: boolean;
  ErrMsg: string;
};

type UpdateUsernameProps = {
  username: string;
};

export default function UpdateUsername(props: UpdateUsernameProps) {
  const router = useRouter();
  const auth = useUser();

  const nameData = trpc.users.getUsername.useMutation();
  const userData = trpc.users.getAuthenticated.useQuery();
  const formData = trpc.users.update.useMutation();
  const [saveSucess, setSaveSucess] = useState(false);
  const [saveError, setSaveError] = useState(false);
  const [formSubmit, setFormSubmit] = useState(false);

  const [username, setUsername] = useState(props.username);
  const [disableSaveUsername, setDisableSaveUsername] = useState(true);
  const [input, setInput] = useState({
    Err: false,
    ErrMsg: '',
  } as InputValidation);

  useEffect(() => {
    if (username == '' && userData?.data?.username) {
      setUsername(userData.data.username);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [userData]);

  // handle formData mutation effects
  useEffect(() => {
    const { isLoading, isError, isSuccess, data } = formData;

    if (!formSubmit || isLoading) {
      return;
    }

    // prevent infinite loop
    setFormSubmit(false);

    if (isSuccess) {
      setSaveError(false);
      setSaveSucess(true);
      router.refresh();
      return;
    }

    // error
    setSaveSucess(false);
    setSaveError(true);

    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [formData, formSubmit]);

  const Alert = React.forwardRef<HTMLDivElement, AlertProps>(function Alert(
    props,
    ref
  ) {
    return <MuiAlert elevation={6} ref={ref} variant="filled" {...props} />;
  });

  // input validation: username
  useEffect(() => {
    // return if init
    if (!userData?.data) return;

    // return if no change
    if (username == userData?.data?.username) {
      setInput({
        Err: false,
        ErrMsg: '',
      } as InputValidation);
      setDisableSaveUsername(true);
      return;
    }
    // client-side validation
    const re = /^[a-zA-Z0-9][a-zA-Z0-9-]{6,22}[a-zA-Z0-9]$/i;
    const { success } = z.string().regex(re).safeParse(username);
    if (!success || username.includes('--')) {
      // handle error
      setDisableSaveUsername(true);
      setInput({
        Err: true,
        ErrMsg:
          'Invalid Username. Min 8 char. Max 24. AlphaNumeric and hyphen, but no "--". (c)',
      } as InputValidation);
      return;
    }
    setDisableSaveUsername(false);
    setInput({ Err: false, ErrMsg: '' } as InputValidation);

    // trigger server-side validation
    nameData.mutate({ username: username });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [username]);

  function handleUsernameChange(e: React.ChangeEvent<HTMLInputElement>) {
    setUsername(e.target.value);
  }

  function handleUpdateUsername() {
    formData.mutate({ username: username });
    setFormSubmit(true);
  }

  function handleClose(event?: React.SyntheticEvent | Event, reason?: string) {
    if (reason === 'clickaway') {
      return;
    }

    setSaveSucess(false);
    setSaveError(false);
  }

  if (auth.isLoading)
    return (
      <Container maxWidth="sm" style={{ paddingTop: 40 }}>
        <Box
          sx={{
            marginTop: 8,
            display: 'flex',
            flexDirection: 'column',
            alignItems: 'center',
          }}
        >
          <CircularProgress />
        </Box>
      </Container>
    );

  if (auth?.error) return <>unexpected error: {auth?.error?.message}</>;

  return (
    <>
      <Snackbar open={saveSucess} autoHideDuration={2000} onClose={handleClose}>
        <Alert onClose={handleClose} severity="success">
          Save successful!
        </Alert>
      </Snackbar>
      <TextField
        fullWidth
        name="username"
        label={'Username'}
        type="username"
        id="username"
        autoComplete="new-username"
        onChange={handleUsernameChange}
        error={input.Err}
        helperText={input.ErrMsg}
        value={username}
      />
      <Typography variant="subtitle2" component="h2">
        <b>Renaming may take a few minutes to complete.</b>
      </Typography>
      <Button
        variant="contained"
        disabled={disableSaveUsername}
        onClick={handleUpdateUsername}
      >
        Save
      </Button>
    </>
  );
}
