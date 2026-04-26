'use client';
import React, { useCallback, useEffect, useState } from 'react';
import { useFilePicker, FileContent } from 'use-file-picker';
import { useRecipeContext } from '@/providers/recipe-context';
import { RecipeMediaViewImage } from '@/components/Recipe/RecipeMedia/RecipeMediaViewImage';
import Grid from '@mui/material/Grid2';
import Button from '@mui/material/Button';
import Box from '@mui/material/Box';
import Badge from '@mui/material/Badge';
import CloudUploadIcon from '@mui/icons-material/CloudUpload';
import log from '@/libs/logger';
import { trpc } from '@/trpc/client';
import { RecipeMediaProps } from '@/types/recipe';
import Stack from '@mui/material/Stack';
import CircularProgress from '@mui/material/CircularProgress';
import { green } from '@mui/material/colors';

export default function RecipeMedia({ user, recipe, media }: RecipeMediaProps) {
  const rtx = useRecipeContext();
  // const user = trpc.users.getAuthenticated.useQuery().data;
  // const signedURLData = trpc.recipes.getSignedURL.useMutation();
  const signedURLData = trpc.recipes.getSignedURL.useMutation();

  const [fetchingSignedURL, setFetchingSignedURL] = useState(false);

  const [openFileSelector, { filesContent, loading, errors, clear }] =
    useFilePicker({
      readAs: 'DataURL',
      accept: 'image/*',
      multiple: false,
      limitFilesConfig: { max: 2 },
      // minFileSize: 1,
      maxFileSize: 5, // in megabytes
    });

  useEffect(() => {
    // log().Debug(new Error(), [{ k: 'msg', v: 'reset media' }]);
    rtx.resetMedia();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const uploadFile = useCallback(
    async (signedURL: string) => {
      if (!signedURL) {
        log().Error(new Error(), [{ k: 'msg', v: 'signedURL is undefined' }]);
        alert('unexpected error. invalid signed url');
        return;
      }

      const file = filesContent[0];
      const ct = getContentTypeFromFileExt(file.name);
      if (!ct) {
        log().Error(new Error(), [
          { k: 'msg', v: 'invalid content type' },
          { k: 'filename', v: file.name },
        ]);
        alert('unexpected error. failed to get content type');
        return;
      }

      // https://stackoverflow.com/questions/59836220/how-to-get-the-equivalent-data-format-of-curl-t-upload-data-option-from-inpu
      const r = await fetch(file.content);
      const blob = await r.blob();
      const buf = await blob.arrayBuffer();

      try {
        await fetch(signedURL, {
          method: 'PUT',
          headers: new Headers({ 'Content-Type': ct }),
          body: buf,
        });
        // tr@ck: UX user feedback
        log().Debug(new Error(), [{ k: 'msg', v: 'upload successful' }]);
        clear();
        // tr@ck: better way to refresh media after allowing time for cloud image proessing?
        await new Promise((r) => setTimeout(r, 2000));
        rtx.resetMedia();
      } catch (e) {
        log().Error(new Error(), [
          { k: 'msg', v: 'failed to upload' },
          { k: 'error: ', v: (e as any).toString() },
        ]);
      }
    },
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [filesContent]
  );

  useEffect(() => {
    const { isPending, isError, isSuccess, data } = signedURLData;

    if (isPending || !fetchingSignedURL) {
      return;
    }

    setFetchingSignedURL(false);

    if (isError || !data?.signedURL || !isSuccess) {
      rtx.setActionInProgress(false);
      const msg = 'unexpected error. failed to get signed url';
      log().Error(new Error(), [{ k: 'msg', v: msg }]);
      alert(msg);
      return;
    }

    try {
      uploadFile(data.signedURL);
    } catch (err) {
      const msg = 'unexpected error. invalid signed url';
      log().Error(new Error(), [
        { k: 'msg', v: msg },
        { k: 'error', v: err as string },
      ]);
      alert(msg);
    } finally {
      rtx.setActionInProgress(false);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [signedURLData, fetchingSignedURL, uploadFile]);

  // async function putMedia(file: FileContent) {}

  function getContentTypeFromFileExt(filename: string): string | undefined {
    var ext = filename.split('.').pop() || '';
    if (ext == 'png') {
      return 'image/png';
    } else if (['jpeg', 'jpg'].includes(ext)) {
      return 'image/jpeg';
    } else {
      console.error('invalid file type');
    }
    return undefined;
  }

  if (loading) {
    return (
      <Box
        component="main"
        sx={{
          flexGrow: 1,
          bgcolor: 'background.default',
          mt: ['2px', '2px', '2px'],
          p: 3,
        }}
      >
        <CircularProgress color="secondary" />
      </Box>
    );
  }

  if (errors.length) {
    return <div>Unexpected Error</div>;
  }

  async function handleUploadMedia() {
    // tr@ck: handle errors ux
    if (!filesContent || filesContent.length != 1) {
      return;
    }

    const filename = filesContent[0].name;
    const ct = getContentTypeFromFileExt(filename);
    if (!ct) {
      alert('invalid file type!');
      return;
    }
    rtx.setActionInProgress(true);
    setFetchingSignedURL(true);
    signedURLData.mutate({
      recipeID: recipe?.id || rtx?.recipeId,
      payload: { filename: filename },
    });
  }

  function handleSelectMedia() {
    openFileSelector();
  }

  const isContributor =
    recipe?.contributors &&
    recipe?.contributors.map((c) => c.id).includes(user?.id);

  function SelectMediaButton() {
    return (
      <Button
        startIcon={<CloudUploadIcon />}
        variant="outlined"
        onClick={handleSelectMedia}
        sx={{ marginBottom: 2 }}
      >
        Select Image
      </Button>
    );
  }

  function UploadMediaButton() {
    return (
      <>
        <Button
          startIcon={<CloudUploadIcon />}
          variant="outlined"
          onClick={handleUploadMedia}
          sx={{ marginBottom: 2 }}
        >
          Upload Image
        </Button>
        {loading && (
          <CircularProgress
            size={24}
            sx={{
              color: green[500],
              position: 'absolute',
              top: '50%',
              left: '50%',
              marginTop: '-12px',
              marginLeft: '-12px',
            }}
          />
        )}
      </>
    );
  }

  function newMediaControls() {
    if (!isContributor) {
      return <>Fork to upload image</>;
    }

    return (
      <Box sx={{ m: 1, position: 'relative' }}>
        {/* <Stack> */}
        {filesContent.length == 0 ? (
          <SelectMediaButton />
        ) : (
          <UploadMediaButton />
        )}

        {filesContent.map((file, index) => {
          return (
            <Grid
              key={index}
              size="grow"
              display="flex"
              justifyContent="center"
              alignItems="center"
            >
              <Badge
                onClick={() => clear()}
                color="primary"
                badgeContent={'x'}
                sx={{ cursor: 'pointer' }}
              >
                <Box
                  component="img"
                  sx={{
                    // height: 256,
                    width: 384,
                    // maxHeight: { xs: 256, md: 256 },
                    maxWidth: { xs: 384, md: 384 },
                    borderStyle: 'solid',
                    border: '2px solid rgb(0, 120, 212)',
                    marginBottom: 2,
                  }}
                  alt={file.name}
                  src={file.content}
                />
              </Badge>
            </Grid>
          );
        })}
      </Box>
    );
  }

  return (
    <Stack>
      {!!user?.id ? newMediaControls() : <>Login to upload images</>}
      <Grid container spacing={1}>
        {rtx.media.map((m) => (
          <RecipeMediaViewImage key={m.id} media={m} />
        ))}
      </Grid>
    </Stack>
  );
}
