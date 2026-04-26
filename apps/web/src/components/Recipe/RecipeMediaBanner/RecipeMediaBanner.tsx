'use client';
import React, { useState, useEffect, useMemo } from 'react';
import { useRecipeContext } from '@/providers/recipe-context';
import { models_RecipeMediaVariant } from '@4ks/api-fetch';
import DialogActions from '@mui/material/DialogActions';
import DialogContent from '@mui/material/DialogContent';
import DialogTitle from '@mui/material/DialogTitle';
import Grid from '@mui/material/Grid2';
import Dialog from '@mui/material/Dialog';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Skeleton from '@mui/material/Skeleton';
import IconButton from '@mui/material/IconButton';
import CloseIcon from '@mui/icons-material/Close';
import ExpandMoreIcon from '@mui/icons-material/ExpandMore';
import ExpandLessIcon from '@mui/icons-material/ExpandLess';
import RecipeMediaBannerImagePreview from './RecipeMediaBannerImagePreview';
import { unstable_noStore as noStore } from 'next/cache';
import { getRecipeBannerVariantUrl } from '@/libs/media';

interface RecipeMediaBannerProps {
  isNew: boolean;
}

export default function RecipeMediaBanner({ isNew }: RecipeMediaBannerProps) {
  noStore();
  const rtx = useRecipeContext();

  const fallbackMediaURL = process.env.MEDIA_FALLBACK_URL;

  const [imageSrc, setImageSrc] = useState<string>();
  const [showBanner, setShowBanner] = useState(true);
  const [bannerImageReady, setBannerImageReady] = useState(false);
  const [showBannerSelectModal, setShowBannerSelectModal] = useState(false);

  const [selectingMediaId, setSelectingMediaId] = useState<string>();
  const [selectingMedia, setSelectingMedia] =
    useState<models_RecipeMediaVariant[]>();

  const random = useMemo(() => Math.floor(Math.random() * 27), []);

  function setRandomImage() {
    setImageSrc(`${fallbackMediaURL}/f${random}.jpg`);
  }

  useEffect(() => {
    const b = localStorage.getItem('showBanner') || 'true';
    setShowBanner(/true/i.test(b));
  }, []);

  useEffect(() => {
    if (rtx?.recipe?.currentRevision?.banner) {
      const md = getRecipeBannerVariantUrl(rtx.recipe.currentRevision.banner);
      md?.url && typeof md.url == 'string' && setImageSrc(md.url);
    } else {
      setRandomImage();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [rtx]);

  useEffect(() => {
    if (rtx?.recipe?.currentRevision?.banner) {
      const bannerImg = getRecipeBannerVariantUrl(
        rtx.recipe.currentRevision.banner
      );
      if (showBannerSelectModal) {
        const newBannerImg = getRecipeBannerVariantUrl(selectingMedia);
        setImageSrc(newBannerImg?.url || bannerImg?.url);
      } else {
        setImageSrc(bannerImg?.url || undefined);
      }
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [showBannerSelectModal, selectingMedia]);

  // tr@ck: save currently happens in RecipeControls. Fix flow.
  function confirmImageSelection() {
    selectingMedia && rtx.setBanner(selectingMedia);
    setShowBannerSelectModal(false);
  }

  function discardImageSelection() {
    setSelectingMedia(undefined);
    setSelectingMediaId(undefined);
    setShowBannerSelectModal(false);
  }

  function handleOpenBannerSelectModal() {
    if (rtx.media && rtx.media.length > 0) {
      setShowBannerSelectModal(true);
    }
  }

  function toggleBanner() {
    const b = !showBanner;
    setShowBanner(b);
    localStorage.setItem('showBanner', b ? 'true' : 'false');
  }

  return (
    <>
      <Dialog
        fullWidth={true}
        maxWidth="xl"
        open={showBannerSelectModal}
        onClose={discardImageSelection}
      >
        <DialogTitle sx={{ m: 0, p: 2 }}>
          Select Banner Image
          <IconButton
            aria-label="close"
            onClick={discardImageSelection}
            sx={{
              position: 'absolute',
              right: 8,
              top: 8,
            }}
          >
            <CloseIcon />
          </IconButton>
        </DialogTitle>
        <DialogContent dividers>
          <Grid container spacing={1}>
            {rtx.media.map((m) =>
              RecipeMediaBannerImagePreview({
                media: m,
                selectingMediaId,
                setSelectingMedia,
                setSelectingMediaId,
              })
            )}
          </Grid>
        </DialogContent>
        <DialogActions>
          <Button onClick={confirmImageSelection}>Select</Button>
          <span style={{ marginLeft: '8px' }}></span>
          <Button onClick={discardImageSelection}>Cancel</Button>
        </DialogActions>
      </Dialog>

      {/* Banner Image */}
      {showBanner && (
        <>
          {/* <Badge
            color="primary"
            badgeContent={<ExpandLessIcon />}
            onClick={toggleBanner}
          > */}
          <Box
            component="img"
            onClick={handleOpenBannerSelectModal}
            sx={{
              width: '100%',
              maxHeight: '512px',
              objectFit: 'cover',
              overflow: 'hidden',
              display: bannerImageReady ? 'block' : 'none',
            }}
            alt="banner image"
            src={imageSrc}
            onLoad={() => {
              setBannerImageReady(true);
            }}
          />
          {/* </Badge> */}
          {!bannerImageReady && <Skeleton variant="rectangular" height={512} />}
        </>
      )}
      <Button
        variant="text"
        onClick={toggleBanner}
        sx={{
          paddingBottom: 2,
          marginBottom: 2,
          width: '100%',
          backgroundColor: '#f6fafd',
          '&:hover': {
            backgroundColor: '#ffffff',
          },
        }}
      >
        {showBanner ? <ExpandLessIcon /> : <ExpandMoreIcon />}
      </Button>
    </>
  );
}
