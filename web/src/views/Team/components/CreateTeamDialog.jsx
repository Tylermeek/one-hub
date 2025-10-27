import { useState } from 'react';
import PropTypes from 'prop-types';
import { Users } from 'lucide-react';
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import { API } from 'utils/api';
import { showError, showSuccess } from 'utils/common';

const CreateTeamDialog = ({ open, onOpenChange, onSuccess }) => {
  const [loading, setLoading] = useState(false);
  const [teamData, setTeamData] = useState({
    name: '',
    description: ''
  });

  const handleSubmit = async (e) => {
    e.preventDefault();

    if (!teamData.name.trim()) {
      showError('请输入团队名称');
      return;
    }

    setLoading(true);
    try {
      const response = await API.post('/api/team/', {
        name: teamData.name.trim(),
        description: teamData.description.trim()
      });

      if (response.data.success) {
        showSuccess('团队创建成功');
        setTeamData({ name: '', description: '' });
        onOpenChange(false);
        if (onSuccess) onSuccess();
      } else {
        showError(response.data.message || '创建团队失败');
      }
    } catch (error) {
      console.error('创建团队失败:', error);
      showError('创建团队失败');
    } finally {
      setLoading(false);
    }
  };

  const handleClose = () => {
    if (!loading) {
      setTeamData({ name: '', description: '' });
      onOpenChange(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[500px]">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Users className="w-5 h-5" />
            创建团队
          </DialogTitle>
          <DialogDescription>创建一个新团队来管理成员和共享额度</DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit}>
          <div className="space-y-4 py-4">
            <div className="space-y-2">
              <Label htmlFor="team-name">
                团队名称 <span className="text-destructive">*</span>
              </Label>
              <Input
                id="team-name"
                placeholder="请输入团队名称"
                value={teamData.name}
                onChange={(e) => setTeamData({ ...teamData, name: e.target.value })}
                maxLength={50}
                disabled={loading}
                autoFocus
              />
              <p className="text-xs text-muted-foreground">{teamData.name.length}/50 字符</p>
            </div>

            <div className="space-y-2">
              <Label htmlFor="team-description">团队描述</Label>
              <Textarea
                id="team-description"
                placeholder="请输入团队描述（可选）"
                value={teamData.description}
                onChange={(e) => setTeamData({ ...teamData, description: e.target.value })}
                maxLength={200}
                rows={3}
                disabled={loading}
              />
              <p className="text-xs text-muted-foreground">{teamData.description.length}/200 字符</p>
            </div>
          </div>

          <DialogFooter>
            <Button type="button" variant="outline" onClick={handleClose} disabled={loading}>
              取消
            </Button>
            <Button type="submit" disabled={loading || !teamData.name.trim()}>
              {loading ? '创建中...' : '创建团队'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
};

CreateTeamDialog.propTypes = {
  open: PropTypes.bool.isRequired,
  onOpenChange: PropTypes.func.isRequired,
  onSuccess: PropTypes.func
};

export default CreateTeamDialog;
