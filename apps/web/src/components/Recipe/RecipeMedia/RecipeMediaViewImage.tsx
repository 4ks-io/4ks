'use client';
import React, { useEffect, useState, useMemo } from 'react';
import { unstable_noStore as noStore } from 'next/cache';
import { models_MediaStatus, models_RecipeMedia } from '@4ks/api-fetch';
import { RecipeMediaSize } from '@/libs/media';
import Box from '@mui/material/Box';
import Stack from '@mui/material/Stack';
import Skeleton from '@mui/material/Skeleton';
import Grid from '@mui/material/Grid2';
import Typography from '@mui/material/Typography';
import { useShiftHoldReveal } from '@/libs/use-shift-hold-reveal';

interface RecipeMediaViewImageProps {
  media: models_RecipeMedia;
}

const MEDIA_SOURCE_AI = 1;

function getMediaSource(media: models_RecipeMedia) {
  return (media as models_RecipeMedia & { source?: number }).source;
}

export function RecipeMediaViewImage({ media }: RecipeMediaViewImageProps) {
  noStore();
  const showDiagnostics = useShiftHoldReveal();
  const [imageSrc, setImageSrc] = useState<string>();
  const [filename, setFilename] = useState('unknown');
  const [imageLoadFailed, setImageLoadFailed] = useState(false);

  useEffect(() => {
    setImageLoadFailed(false);
    if (media.variants && media.variants.length > 0) {
      let sm = media.variants.filter((v) => v.alias == RecipeMediaSize.SM)[0];
      if (sm) {
        setImageSrc(sm.url);
        setFilename(`${sm.filename}`);
      }
    } else {
      setRandomImage();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [media.variants]);

  const random = useMemo(() => Math.floor(Math.random() * 27), []);

  function setRandomImage() {
    setImageSrc(`${process.env.MEDIA_FALLBACK_URL}/f${random}.jpg`);
  }

  const mediaSource = getMediaSource(media);
  const isGeneratingAIImage =
    mediaSource === MEDIA_SOURCE_AI &&
    (media.status === models_MediaStatus.MediaStatusRequested ||
      media.status === models_MediaStatus.MediaStatusProcessing);
  const isUploadingImage =
    mediaSource !== MEDIA_SOURCE_AI &&
    (media.status === models_MediaStatus.MediaStatusRequested ||
      media.status === models_MediaStatus.MediaStatusProcessing);
  const placeholderLabel = isGeneratingAIImage
    ? 'Generating AI image'
    : isUploadingImage
      ? 'Upload in progress'
    : undefined;
  const diagnosticLines = [
    `id: ${media.id ?? 'unknown'}`,
    `source: ${mediaSource ?? 'unknown'}`,
    `status: ${media.status ?? 'unknown'}`,
    `recipe: ${media.recipeId ?? 'unknown'}`,
    `root: ${media.rootRecipeId ?? 'unknown'}`,
    `file: ${filename}`,
    `url: ${imageSrc ?? 'unknown'}`,
  ];

  // function handleError(loadState: ImageLoadState) {
  //   if (loadState == ImageLoadState.error) {
  //     setRandomImage();
  //   }
  // }

  return (
    <Grid size={{ xs: 12, md: 6, lg: 4 }} key={media.id} sx={{ pb: 3 }}>
      <Stack direction="row" justifyContent="center" alignItems="center">
        {imageSrc && !imageLoadFailed ? (
          <Box
            component="img"
            sx={{
              width: '100%',
              height: 'auto',
              maxWidth: 384,
              display: 'block',
            }}
            alt={filename}
            src={imageSrc}
            onError={() => setImageLoadFailed(true)}
          />
        ) : !imageLoadFailed && (placeholderLabel || showDiagnostics) ? (
          <Box sx={{ position: 'relative', width: 384, maxWidth: '100%' }}>
            <Skeleton variant="rectangular" width="100%" height={256} />
            {(placeholderLabel || showDiagnostics) && (
              <Box
                sx={{
                  position: 'absolute',
                  inset: 0,
                  display: 'flex',
                  flexDirection: 'column',
                  alignItems: 'center',
                  justifyContent: 'center',
                  px: 2,
                  color: 'text.secondary',
                }}
              >
                {placeholderLabel && (
                  <Typography variant="button" sx={{ letterSpacing: 0 }}>
                    {placeholderLabel}
                  </Typography>
                )}
                {showDiagnostics && (
                  <Box
                    component="pre"
                    sx={{
                      width: '100%',
                      maxHeight: 150,
                      mt: placeholderLabel ? 1 : 0,
                      mb: 0,
                      overflow: 'hidden',
                      whiteSpace: 'pre-wrap',
                      overflowWrap: 'anywhere',
                      fontFamily: 'monospace',
                      fontSize: 11,
                      lineHeight: 1.35,
                      textAlign: 'left',
                    }}
                  >
                    {diagnosticLines.join('\n')}
                  </Box>
                )}
              </Box>
            )}
          </Box>
        ) : null}
      </Stack>
    </Grid>
  );

  // return (
  //   <Image
  //     key={media.id}
  //     src={imageSrc}
  //     onLoadingStateChange={handleError}
  //     imageFit={ImageFit.cover}
  //     alt={filename}
  //     width={256}
  //     height={160}
  //   />
  // );
}
