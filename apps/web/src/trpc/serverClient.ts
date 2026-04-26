import { appRouter } from '@/server';

export const serverClient = appRouter.createCaller({});
