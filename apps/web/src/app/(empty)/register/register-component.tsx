'use client';
import React, { useEffect, useState } from 'react';
import { trpc } from '@/trpc/client';
import { z } from 'zod';
import { redirect } from 'next/navigation';
import Image from 'next/image';
import Avatar from '@mui/material/Avatar';
import Button from '@mui/material/Button';
import CssBaseline from '@mui/material/CssBaseline';
import TextField from '@mui/material/TextField';
import Grid from '@mui/material/Grid';
import Box from '@mui/material/Box';
import AccountCircleIcon from '@mui/icons-material/AccountCircle';
import Typography from '@mui/material/Typography';
import Container from '@mui/material/Container';
import UsernameSpecification from '@/components/UsernameSpecifications';
import type { User } from '@auth0/nextjs-auth0/types';

type formValidationError = {
  validation: string;
  code: string;
  message: string;
  path: string[];
};

type InputValidation = {
  Err: boolean;
  ErrMsg: string;
};

type Input = {
  username: InputValidation;
  email: InputValidation;
  displayName: InputValidation;
};

function NewInputMap() {
  return {
    email: { Err: false, ErrMsg: '' },
    username: { Err: false, ErrMsg: '' },
    displayName: { Err: false, ErrMsg: '' },
  } as Input;
}

type RegisterComponentProps = {
  user: User;
};

export default function RegisterComponent({ user }: RegisterComponentProps) {
  // const { user, error, isLoading } = useUser();

  // validation hooks
  const [hasError, setHasError] = useState(false);
  const [input, setInput] = useState<Input>(NewInputMap());
  const [isNewFormError, setIsNewFormError] = useState(true);

  // input hooks
  const [displayName, setDisplayName] = useState('');
  const [email, setEmail] = useState(user?.email || '');
  const [username, setUsername] = useState('');

  // data hooks
  const formData = trpc.users.create.useMutation();
  const nameData = trpc.users.getUsername.useMutation();

  // initial user details
  useEffect(() => {
    const names = user?.name?.split(' ');
    if (names) {
      setDisplayName(`${names[0]}` || '');
    }
    setEmail(`${user?.email}` || '');
    setUsername(formatUsername(user?.nickname));
  }, [user]);

  // check for input error. allow/prevent form submit
  useEffect(() => {
    let e = false;
    Object.keys(input).map((key) => {
      input[key as keyof Input].Err && (e = true);
    });
    setHasError(e);
  }, [input]);

  // input validation: username
  useEffect(() => {
    // client-side validation
    const re = /^[a-zA-Z0-9][a-zA-Z0-9-]{6,22}[a-zA-Z0-9]$/i;
    const { success } = z.string().regex(re).safeParse(username);
    if (!success || username.includes('--')) {
      // handle error
      setInputError(
        'username',
        "Invalid Username. Min 8 char. Max 24. AlphaNumeric and hyphen, but no '--'. (c)"
      );
      return;
    }
    clearInputError('username');

    // trigger server-side validation
    nameData.mutate({ username: username });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [username]);

  // client-side input validation: email
  useEffect(() => {
    const { success } = z.string().email().safeParse(email);
    if (success) {
      clearInputError('email');
      return;
    }
    // handle error
    setInputError('email', 'Invalid email address (c)');
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [email]);

  function isValidHumanName(name: string): boolean {
    const r0 = /^[a-zA-Z-]{0,47}[a-zA-Z-]$/i;
    const r1 = /^[a-zA-Z]/; // begin alpha
    const r2 = /[a-zA-Z]$/; // end alpha
    const { success } = z
      .string()
      .regex(r0)
      .regex(r1)
      .regex(r2)
      .safeParse(name);

    if (!success || name.includes('--')) {
      return false;
    }
    return true;
  }

  // client-side input validation: displayName
  useEffect(() => {
    if (isValidHumanName(displayName)) {
      clearInputError('displayName');
      return;
    }
    // handle error
    setInputError(
      'displayName',
      'Invalid Display Name. Min 1 char. Max 48. Alphabetic and hyphens only. (c)'
    );
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [displayName]);

  function setInputError(f: keyof Input, msg: string) {
    setInputField(f, { Err: true, ErrMsg: msg });
  }

  function clearInputError(f: keyof Input) {
    setInputField(f, { Err: false, ErrMsg: '' });
  }

  function setInputField(f: keyof Input, v: InputValidation) {
    // eslint-disable-next-line prefer-const
    let i = DeepCopyInput(input);
    i[f] = v;
    safeSetInput(i);
  }

  function DeepCopyInput(i: Input) {
    return JSON.parse(JSON.stringify(i)) as Input;
  }

  // safeSetInput prevents infinite loop by only updating hook if data has truly changed
  function safeSetInput(i: Input) {
    // works thanks to NewInputMap() returning keys in same order
    if (JSON.stringify(i) != JSON.stringify(input)) {
      setInput(i);
    }
  }

  // handle formData mutation effects
  useEffect(() => {
    const { isLoading, isError, isSuccess, data } = formData;
    if (isLoading) {
      return;
    }
    if (isSuccess) {
      redirect('/explore');
    }
    if (!isError) {
      return;
    }
    // prevent infinite loop
    if (!isNewFormError) {
      return;
    }
    setIsNewFormError(false);

    // update progress state
    const msg = formData.error.message;

    // attempt to parse json
    let err;
    try {
      err = JSON.parse(msg);
    } catch (e) {
      console.log('unexpected error', e);
      return;
    }

    // handle unexpected error
    if (!Array.isArray(err)) {
      console.log('unexpected error', err);
      return;
    }

    // aggregate errors
    // eslint-disable-next-line prefer-const
    let i = NewInputMap();
    err.map((e: formValidationError) => {
      e.path.map((p: string) => {
        i[p as keyof Input] = {
          ErrMsg: newCSV(i[p as keyof Input].ErrMsg, e.message),
          Err: true,
        } as InputValidation;
      });
    });

    // set
    safeSetInput(i);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [formData]);

  // handle username query mutation effects
  useEffect(() => {
    const { isLoading, isError, data } = nameData;
    // ignore previous queries
    if (data?.username != username) {
      return;
    }
    if (isLoading) {
      return;
    }
    if (isError || !data) {
      setInputError('username', 'Unknown error');
      return;
    }
    if (!data?.valid) {
      // tr@ck: load reserved list from server and perform check client-side
      if (data.msg == 'reserved') {
        setInputError(
          'username',
          `'${username}' is a reserved username. Please choose another.`
        );
        return;
      }
      setInputError(
        'username',
        "Invalid Username. Min 8 char. Max 24. AlphaNumeric and hyphen, but no '--'."
      );
      return;
    }
    if (!data?.available) {
      setInputError('username', 'Username already in use.');
      return;
    }

    clearInputError('username');
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [nameData]);

  // newCSV will return or append strings
  function newCSV(msg: string, append: string) {
    if (msg == '') {
      return append;
    }
    return `${msg}, ${append}`;
  }

  // helper
  function formatUsername(username: string | null | undefined) {
    let u = `${username}`;
    u = u.replaceAll('_', '-');
    u = u.replaceAll('.', '-');
    return u;
  }

  function handleUsernameChange(e: React.ChangeEvent<HTMLInputElement>) {
    setUsername(e.target.value);
  }

  function handleEmailChange(e: React.ChangeEvent<HTMLInputElement>) {
    setEmail(e.target.value);
  }

  function handleDisplayNameChange(e: React.ChangeEvent<HTMLInputElement>) {
    setDisplayName(e.target.value);
  }

  async function handleRegistration() {
    setIsNewFormError(true);
    formData.mutate({
      username: username,
      email: email,
      displayName: displayName,
    });
  }

  return (
    <>
      <Container component="main" maxWidth="xs">
        <CssBaseline />
        <Box
          sx={{
            marginTop: 8,
            display: 'flex',
            flexDirection: 'column',
            alignItems: 'center',
          }}
        >
          <Avatar sx={{ m: 1, bgcolor: 'secondary.main' }}>
            <AccountCircleIcon />
          </Avatar>
          <Typography component="h1" variant="h5">
            New User Registration
          </Typography>
          {/* <Image
            alt="Avatar"
            src={`${user?.picture}`}
            width={96}
            height={96}
            priority={false}
          /> */}
          <Box sx={{ mt: 3 }}>
            <Grid container spacing={2}>
              <Grid item xs={12} sm={6}>
                <TextField
                  autoComplete="display-name"
                  name="displayName"
                  required
                  fullWidth
                  id="displayName"
                  label={'Display Name'}
                  autoFocus
                  value={displayName}
                  onChange={handleDisplayNameChange}
                  error={input.displayName.Err}
                  helperText={input.displayName.ErrMsg}
                />
              </Grid>
              <Grid item xs={12}>
                <TextField
                  required
                  fullWidth
                  id="email"
                  label={'Email Address'}
                  name="email"
                  autoComplete="email"
                  value={email}
                  onChange={handleEmailChange}
                  error={input.email.Err}
                  helperText={input.email.ErrMsg}
                />
              </Grid>
              <Grid item xs={12}>
                <TextField
                  required
                  fullWidth
                  name="username"
                  label={'Username'}
                  type="username"
                  id="username"
                  autoComplete="new-username"
                  onChange={handleUsernameChange}
                  error={input.username.Err}
                  helperText={input.username.ErrMsg}
                  value={username}
                />
                <UsernameSpecification />
              </Grid>
            </Grid>
            <Button
              fullWidth
              variant="contained"
              sx={{ mt: 3, mb: 2 }}
              onClick={handleRegistration}
              type="submit"
              disabled={hasError}
            >
              Register
            </Button>
          </Box>
        </Box>
      </Container>
    </>
  );
}
