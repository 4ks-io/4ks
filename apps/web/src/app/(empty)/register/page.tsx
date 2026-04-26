import React from 'react';
import RegisterComponent from './register-component';
import { Page, PageProps } from '@/libs/navigation';
import { handleUserNavigation } from '@/libs/server/navigation';
import type { Metadata } from 'next';
import { getSession } from '@auth0/nextjs-auth0';

export const metadata: Metadata = {
  title: '4ks Register',
  description: '4ks User Registration',
};

export default async function RegisterPage({
}: PageProps) {
  const { session } = await handleUserNavigation(Page.REGISTER);

  if (!session || !session?.user) {
    return <div>Error</div>;
  }

  return <RegisterComponent user={session.user} />;
}
