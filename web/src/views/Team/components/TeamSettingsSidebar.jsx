import React from 'react';
import PropTypes from 'prop-types';
import { useTranslation } from 'react-i18next';
import { Info, DollarSign, Shield, Palette, AlertTriangle } from 'lucide-react';
import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/button';

/**
 * 设置导航项配置
 */
const getNavigationItems = (t, isOwner) => [
  {
    id: 'basic',
    label: t('team_settings.nav.basic_info'),
    icon: Info,
    description: t('team_settings.nav.basic_info_desc'),
    visible: true
  },
  {
    id: 'quota',
    label: t('team_settings.nav.quota_management'),
    icon: DollarSign,
    description: t('team_settings.nav.quota_management_desc'),
    visible: isOwner
  },
  {
    id: 'permissions',
    label: t('team_settings.nav.permissions'),
    icon: Shield,
    description: t('team_settings.nav.permissions_desc'),
    visible: isOwner
  },
  {
    id: 'customization',
    label: t('team_settings.nav.customization'),
    icon: Palette,
    description: t('team_settings.nav.customization_desc'),
    visible: true
  },
  {
    id: 'danger',
    label: t('team_settings.nav.danger_zone'),
    icon: AlertTriangle,
    description: t('team_settings.nav.danger_zone_desc'),
    visible: isOwner
  }
];

/**
 * 团队设置侧边栏导航组件
 */
function TeamSettingsSidebar({ activeSection, onSectionChange, isOwner, className }) {
  const { t } = useTranslation();
  const navigationItems = getNavigationItems(t, isOwner).filter((item) => item.visible);

  return (
    <nav className={cn('space-y-1', className)}>
      {navigationItems.map((item) => {
        const Icon = item.icon;
        const isActive = activeSection === item.id;

        return (
          <Button
            key={item.id}
            variant={isActive ? 'secondary' : 'ghost'}
            className={cn('w-full justify-start text-left h-auto py-3 px-4', isActive && 'bg-secondary')}
            onClick={() => onSectionChange(item.id)}
          >
            <div className="flex items-start gap-3 w-full">
              <Icon
                className={cn(
                  'h-5 w-5 flex-shrink-0 mt-0.5',
                  isActive ? 'text-primary' : 'text-muted-foreground',
                  item.id === 'danger' && 'text-destructive'
                )}
              />
              <div className="flex-1 min-w-0">
                <div
                  className={cn(
                    'font-medium text-sm',
                    isActive ? 'text-foreground' : 'text-muted-foreground',
                    item.id === 'danger' && 'text-destructive'
                  )}
                >
                  {item.label}
                </div>
                <div className="text-xs text-muted-foreground mt-0.5 line-clamp-2">{item.description}</div>
              </div>
            </div>
          </Button>
        );
      })}
    </nav>
  );
}

TeamSettingsSidebar.propTypes = {
  activeSection: PropTypes.string.isRequired,
  onSectionChange: PropTypes.func.isRequired,
  isOwner: PropTypes.bool.isRequired,
  className: PropTypes.string
};

export default TeamSettingsSidebar;
