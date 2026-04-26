export type MutationStateSnapshot = {
  isError: boolean;
  isPending: boolean;
  isSuccess: boolean;
};

export function isSettledMutation({
  isError,
  isPending,
  isSuccess,
}: MutationStateSnapshot) {
  return !isPending && (isError || isSuccess);
}

export function shouldHandleSettledMutation(
  enabled: boolean,
  mutation: MutationStateSnapshot
) {
  return enabled && isSettledMutation(mutation);
}
