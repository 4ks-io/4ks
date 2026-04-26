'use client';
import * as React from 'react';
import Button from '@mui/material/Button';
import { navigateToLogin } from '@/libs/navigation';

export default function AppHeaderAvatarUnauthenticated() {
  return <Button onClick={() => navigateToLogin()}>Login</Button>;
}
