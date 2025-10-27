import React, { useState } from 'react';
import PropTypes from 'prop-types';
import { useTranslation } from 'react-i18next';
import { Save, AlertCircle, Loader2, Trash2, ArrowRightLeft, Power, Check } from 'lucide-react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import { Switch } from '@/components/ui/switch';
import { Badge } from '@/components/ui/badge';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { cn } from '@/lib/utils';
import { API } from 'utils/api';
import { showError, showSuccess } from 'utils/common';

/**
 * 基本信息表单
 */
export function BasicInfoForm({ team, onUpdate, isOwner }) {
  const { t } = useTranslation();
  const [isLoading, setIsLoading] = useState(false);
  const [formData, setFormData] = useState({
    name: team.name || '',
    description: team.description || '',
    status: team.status || 1
  });

  const handleSubmit = async (e) => {
    e.preventDefault();
    setIsLoading(true);

    try {
      const response = await API.put(`/api/team/${team.id}`, formData);
      const { success, message } = response.data;

      if (success) {
        showSuccess(t('team_settings.update_success'));
        onUpdate && onUpdate(formData);
      } else {
        showError(message);
      }
    } catch (error) {
      console.error('Failed to update team:', error);
      showError(t('team_settings.update_failed'));
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle>{t('team_settings.basic_info')}</CardTitle>
        <CardDescription>{t('team_settings.basic_info_description')}</CardDescription>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleSubmit} className="space-y-6">
          {/* 团队名称 */}
          <div className="space-y-2">
            <Label htmlFor="team-name">{t('team_settings.team_name')}</Label>
            <Input
              id="team-name"
              value={formData.name}
              onChange={(e) => setFormData({ ...formData, name: e.target.value })}
              placeholder={t('team_settings.team_name_placeholder')}
              required
            />
          </div>

          {/* 团队描述 */}
          <div className="space-y-2">
            <Label htmlFor="team-description">{t('team_settings.team_description')}</Label>
            <Textarea
              id="team-description"
              value={formData.description}
              onChange={(e) => setFormData({ ...formData, description: e.target.value })}
              placeholder={t('team_settings.team_description_placeholder')}
              rows={4}
            />
          </div>

          {/* 团队状态（仅 Owner 可见） */}
          {isOwner && (
            <div className="space-y-2">
              <Label>{t('team_settings.team_status')}</Label>
              <div className="flex items-center gap-3">
                <Switch
                  checked={formData.status === 1}
                  onCheckedChange={(checked) => setFormData({ ...formData, status: checked ? 1 : 0 })}
                />
                <div className="flex items-center gap-2">
                  <Power className={cn('h-4 w-4', formData.status === 1 ? 'text-green-500' : 'text-gray-400')} />
                  <span className="text-sm text-muted-foreground">
                    {formData.status === 1 ? t('common.enabled') : t('common.disabled')}
                  </span>
                </div>
              </div>
              <p className="text-xs text-muted-foreground">{t('team_settings.team_status_description')}</p>
            </div>
          )}

          {/* 提交按钮 */}
          <Button type="submit" disabled={isLoading}>
            {isLoading ? (
              <>
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                {t('common.saving')}
              </>
            ) : (
              <>
                <Save className="mr-2 h-4 w-4" />
                {t('common.save')}
              </>
            )}
          </Button>
        </form>
      </CardContent>
    </Card>
  );
}

BasicInfoForm.propTypes = {
  team: PropTypes.object.isRequired,
  onUpdate: PropTypes.func,
  isOwner: PropTypes.bool.isRequired
};

/**
 * 额度管理表单（仅 Owner）
 */
export function QuotaManagementForm({ team, onUpdate }) {
  const { t } = useTranslation();
  const [isLoading, setIsLoading] = useState(false);
  const [formData, setFormData] = useState({
    quota_limit: team.quota_limit || 0,
    unlimited_quota: team.unlimited_quota || false
  });

  const handleSubmit = async (e) => {
    e.preventDefault();
    setIsLoading(true);

    try {
      const response = await API.put(`/api/team/${team.id}/settings`, {
        quota_limit: formData.unlimited_quota ? -1 : formData.quota_limit
      });
      const { success, message } = response.data;

      if (success) {
        showSuccess(t('team_settings.update_success'));
        onUpdate && onUpdate(formData);
      } else {
        showError(message);
      }
    } catch (error) {
      console.error('Failed to update quota:', error);
      showError(t('team_settings.update_failed'));
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle>{t('team_settings.quota_management')}</CardTitle>
        <CardDescription>{t('team_settings.quota_management_description')}</CardDescription>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleSubmit} className="space-y-6">
          {/* 无限额度开关 */}
          <div className="space-y-2">
            <div className="flex items-center justify-between">
              <Label>{t('team_settings.unlimited_quota')}</Label>
              <Switch
                checked={formData.unlimited_quota}
                onCheckedChange={(checked) => setFormData({ ...formData, unlimited_quota: checked })}
              />
            </div>
            <p className="text-xs text-muted-foreground">{t('team_settings.unlimited_quota_description')}</p>
          </div>

          {/* 额度上限 */}
          {!formData.unlimited_quota && (
            <div className="space-y-2">
              <Label htmlFor="quota-limit">{t('team_settings.quota_limit')}</Label>
              <Input
                id="quota-limit"
                type="number"
                min="0"
                step="0.01"
                value={formData.quota_limit}
                onChange={(e) => setFormData({ ...formData, quota_limit: parseFloat(e.target.value) || 0 })}
                placeholder="0.00"
                required
              />
              <p className="text-xs text-muted-foreground">{t('team_settings.quota_limit_description')}</p>
            </div>
          )}

          <Button type="submit" disabled={isLoading}>
            {isLoading ? (
              <>
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                {t('common.saving')}
              </>
            ) : (
              <>
                <Save className="mr-2 h-4 w-4" />
                {t('common.save')}
              </>
            )}
          </Button>
        </form>
      </CardContent>
    </Card>
  );
}

QuotaManagementForm.propTypes = {
  team: PropTypes.object.isRequired,
  onUpdate: PropTypes.func
};

/**
 * 权限设置表单（仅 Owner）
 */
export function PermissionsForm({ team, onUpdate }) {
  const { t } = useTranslation();
  const [isLoading, setIsLoading] = useState(false);
  const [formData, setFormData] = useState({
    allow_member_view_stats: team.permissions?.allow_member_view_stats ?? true,
    allow_admin_manage_members: team.permissions?.allow_admin_manage_members ?? true,
    allow_admin_view_all_activities: team.permissions?.allow_admin_view_all_activities ?? true
  });

  const handleSubmit = async (e) => {
    e.preventDefault();
    setIsLoading(true);

    try {
      const response = await API.put(`/api/team/${team.id}/settings`, {
        permissions: formData
      });
      const { success, message } = response.data;

      if (success) {
        showSuccess(t('team_settings.update_success'));
        onUpdate && onUpdate({ permissions: formData });
      } else {
        showError(message);
      }
    } catch (error) {
      console.error('Failed to update permissions:', error);
      showError(t('team_settings.update_failed'));
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle>{t('team_settings.permissions')}</CardTitle>
        <CardDescription>{t('team_settings.permissions_description')}</CardDescription>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleSubmit} className="space-y-6">
          {/* 成员查看统计 */}
          <div className="flex items-center justify-between p-4 border rounded-lg">
            <div className="space-y-1">
              <Label>{t('team_settings.allow_member_view_stats')}</Label>
              <p className="text-xs text-muted-foreground">{t('team_settings.allow_member_view_stats_desc')}</p>
            </div>
            <Switch
              checked={formData.allow_member_view_stats}
              onCheckedChange={(checked) => setFormData({ ...formData, allow_member_view_stats: checked })}
            />
          </div>

          {/* 管理员管理成员 */}
          <div className="flex items-center justify-between p-4 border rounded-lg">
            <div className="space-y-1">
              <Label>{t('team_settings.allow_admin_manage_members')}</Label>
              <p className="text-xs text-muted-foreground">{t('team_settings.allow_admin_manage_members_desc')}</p>
            </div>
            <Switch
              checked={formData.allow_admin_manage_members}
              onCheckedChange={(checked) => setFormData({ ...formData, allow_admin_manage_members: checked })}
            />
          </div>

          {/* 管理员查看所有活动 */}
          <div className="flex items-center justify-between p-4 border rounded-lg">
            <div className="space-y-1">
              <Label>{t('team_settings.allow_admin_view_all_activities')}</Label>
              <p className="text-xs text-muted-foreground">{t('team_settings.allow_admin_view_all_activities_desc')}</p>
            </div>
            <Switch
              checked={formData.allow_admin_view_all_activities}
              onCheckedChange={(checked) => setFormData({ ...formData, allow_admin_view_all_activities: checked })}
            />
          </div>

          <Button type="submit" disabled={isLoading}>
            {isLoading ? (
              <>
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                {t('common.saving')}
              </>
            ) : (
              <>
                <Save className="mr-2 h-4 w-4" />
                {t('common.save')}
              </>
            )}
          </Button>
        </form>
      </CardContent>
    </Card>
  );
}

PermissionsForm.propTypes = {
  team: PropTypes.object.isRequired,
  onUpdate: PropTypes.func
};

/**
 * 危险操作区域（仅 Owner）
 */
export function DangerZoneForm({ team, onAction }) {
  const { t } = useTranslation();
  const [showTransferDialog, setShowTransferDialog] = useState(false);
  const [showDissolveDialog, setShowDissolveDialog] = useState(false);
  const [selectedMember, setSelectedMember] = useState('');
  const [confirmText, setConfirmText] = useState('');
  const [isLoading, setIsLoading] = useState(false);

  const handleTransfer = async () => {
    if (!selectedMember) return;
    setIsLoading(true);

    try {
      const response = await API.post(`/api/team/${team.id}/transfer`, {
        new_owner_id: parseInt(selectedMember)
      });
      const { success, message } = response.data;

      if (success) {
        showSuccess(t('team_settings.transfer_success'));
        setShowTransferDialog(false);
        onAction && onAction('transfer');
      } else {
        showError(message);
      }
    } catch (error) {
      console.error('Failed to transfer team:', error);
      showError(t('team_settings.transfer_failed'));
    } finally {
      setIsLoading(false);
    }
  };

  const handleDissolve = async () => {
    if (confirmText !== team.name) return;
    setIsLoading(true);

    try {
      const response = await API.delete(`/api/team/${team.id}`);
      const { success, message } = response.data;

      if (success) {
        showSuccess(t('team_settings.dissolve_success'));
        setShowDissolveDialog(false);
        onAction && onAction('dissolve');
      } else {
        showError(message);
      }
    } catch (error) {
      console.error('Failed to dissolve team:', error);
      showError(t('team_settings.dissolve_failed'));
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <>
      <Card className="border-destructive">
        <CardHeader>
          <CardTitle className="text-destructive flex items-center gap-2">
            <AlertCircle className="h-5 w-5" />
            {t('team_settings.danger_zone')}
          </CardTitle>
          <CardDescription>{t('team_settings.danger_zone_description')}</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          {/* 转让所有权 */}
          <Alert>
            <ArrowRightLeft className="h-4 w-4" />
            <AlertTitle>{t('team_settings.transfer_ownership')}</AlertTitle>
            <AlertDescription className="mt-2">
              <p className="text-sm mb-3">{t('team_settings.transfer_ownership_desc')}</p>
              <Button variant="outline" onClick={() => setShowTransferDialog(true)}>
                <ArrowRightLeft className="mr-2 h-4 w-4" />
                {t('team_settings.transfer_button')}
              </Button>
            </AlertDescription>
          </Alert>

          {/* 解散团队 */}
          <Alert variant="destructive">
            <Trash2 className="h-4 w-4" />
            <AlertTitle>{t('team_settings.dissolve_team')}</AlertTitle>
            <AlertDescription className="mt-2">
              <p className="text-sm mb-3">{t('team_settings.dissolve_team_desc')}</p>
              <Button variant="destructive" onClick={() => setShowDissolveDialog(true)}>
                <Trash2 className="mr-2 h-4 w-4" />
                {t('team_settings.dissolve_button')}
              </Button>
            </AlertDescription>
          </Alert>
        </CardContent>
      </Card>

      {/* 转让对话框 */}
      <Dialog open={showTransferDialog} onOpenChange={setShowTransferDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t('team_settings.transfer_ownership')}</DialogTitle>
            <DialogDescription>{t('team_settings.transfer_confirmation')}</DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <div className="space-y-2">
              <Label>{t('team_settings.new_owner')}</Label>
              <Select value={selectedMember} onValueChange={setSelectedMember}>
                <SelectTrigger>
                  <SelectValue placeholder={t('team_settings.select_member')} />
                </SelectTrigger>
                <SelectContent>
                  {team.members
                    ?.filter((m) => m.role !== 0)
                    .map((member) => (
                      <SelectItem key={member.id} value={member.id.toString()}>
                        {member.username || member.email}
                      </SelectItem>
                    ))}
                </SelectContent>
              </Select>
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowTransferDialog(false)}>
              {t('common.cancel')}
            </Button>
            <Button onClick={handleTransfer} disabled={!selectedMember || isLoading}>
              {isLoading ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : null}
              {t('common.confirm')}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* 解散对话框 */}
      <Dialog open={showDissolveDialog} onOpenChange={setShowDissolveDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle className="text-destructive">{t('team_settings.dissolve_team')}</DialogTitle>
            <DialogDescription>{t('team_settings.dissolve_warning')}</DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <div className="space-y-2">
              <Label>
                {t('team_settings.type_team_name_to_confirm')}: <Badge>{team.name}</Badge>
              </Label>
              <Input value={confirmText} onChange={(e) => setConfirmText(e.target.value)} placeholder={team.name} />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowDissolveDialog(false)}>
              {t('common.cancel')}
            </Button>
            <Button variant="destructive" onClick={handleDissolve} disabled={confirmText !== team.name || isLoading}>
              {isLoading ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : null}
              {t('team_settings.dissolve_button')}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}

DangerZoneForm.propTypes = {
  team: PropTypes.object.isRequired,
  onAction: PropTypes.func
};
