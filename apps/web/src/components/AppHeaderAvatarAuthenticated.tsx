'use client';
import * as React from 'react';
import { useRouter } from 'next/navigation';
import { authLogoutPath, navigateToAuthRoute } from '@/libs/navigation';
import Menu from '@mui/material/Menu';
import Avatar from '@mui/material/Avatar';
import Tooltip from '@mui/material/Tooltip';
import MenuItem from '@mui/material/MenuItem';
import ListItemIcon from '@mui/material/ListItemIcon';
import AccountCircleIcon from '@mui/icons-material/AccountCircle';
import LogoutIcon from '@mui/icons-material/Logout';
import IconButton from '@mui/material/IconButton';
import SettingsIcon from '@mui/icons-material/Settings';

type AppHeaderAvatarAuthenticatedProps = {
  username: string;
};

export default function AppHeaderAvatarAuthenticated({
  username,
}: AppHeaderAvatarAuthenticatedProps) {
  const router = useRouter();

  const [anchorEl, setAnchorEl] = React.useState<null | HTMLElement>(null);
  const open = Boolean(anchorEl);

  function handleMenuClose() {
    setAnchorEl(null);
  }

  function handleMenuOpen(event: React.MouseEvent<HTMLButtonElement>) {
    setAnchorEl(event.currentTarget);
  }

  function handleProfileClick() {
    router.push(`/${username}`);
    setAnchorEl(null);
  }

  function handleSettingsClick() {
    router.push('/settings');
    setAnchorEl(null);
  }

  function handleLogoutOnClick() {
    setAnchorEl(null);
    navigateToAuthRoute(authLogoutPath);
  }

  function ProfileMenu() {
    return (
      <Menu
        id="basic-menu"
        anchorEl={anchorEl}
        open={open}
        onClose={handleMenuClose}
        MenuListProps={{
          'aria-labelledby': 'basic-button',
        }}
        // sx={{ zIndex: 1001 }}
      >
        <MenuItem onClick={handleProfileClick}>
          <ListItemIcon>
            <AccountCircleIcon fontSize="small" />
          </ListItemIcon>
          Profile
        </MenuItem>
        <MenuItem onClick={handleSettingsClick}>
          <ListItemIcon>
            <SettingsIcon fontSize="small" />
          </ListItemIcon>
          Settings
        </MenuItem>
        {/* {!isTransition && <Divider />} */}
        <MenuItem onClick={handleLogoutOnClick}>
          <ListItemIcon>
            <LogoutIcon fontSize="small" />
          </ListItemIcon>
          Logout
        </MenuItem>
      </Menu>
    );
  }

  return (
    <>
      <Tooltip title="Account">
        <IconButton
          onClick={handleMenuOpen}
          aria-controls={open ? 'account-menu' : undefined}
          aria-haspopup="true"
          aria-expanded={open ? 'true' : undefined}
        >
          <Avatar sx={{ width: 28, height: 28 }}>
            {username.substring(0, 1)}
          </Avatar>
        </IconButton>
      </Tooltip>
      <ProfileMenu />
    </>
  );
}
