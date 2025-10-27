import { useState, useEffect } from 'react';
import PropTypes from 'prop-types';
import { Wallet, AlertCircle } from 'lucide-react';
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Switch } from '@/components/ui/switch';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { API } from 'utils/api';
import { showError, showSuccess, renderQuota } from 'utils/common';

const AllocateQuotaDialog = ({ open, onOpenChange, teamId, teamData, onSuccess }) => {
    const [loading, setLoading] = useState(false);
    const [amount, setAmount] = useState('');
    const [unlimited, setUnlimited] = useState(false);

    // 当对话框打开时，初始化数据
    useEffect(() => {
        if (open && teamData) {
            setUnlimited(teamData.unlimited_quota || false);
            if (!teamData.unlimited_quota && teamData.quota) {
                setAmount((teamData.quota / 1000000).toString());
            } else {
                setAmount('');
            }
        }
    }, [open, teamData]);

    const handleSubmit = async (e) => {
        e.preventDefault();

        if (!unlimited && (!amount || parseFloat(amount) <= 0)) {
            showError('请输入有效的金额');
            return;
        }

        // 检查是否超过个人余额
        if (!unlimited && teamData.owner_quota) {
            const inputAmount = parseFloat(amount);
            const ownerBalance = teamData.owner_quota / 1000000;
            if (inputAmount > ownerBalance) {
                showError(`团队消费上限不能超过您的个人余额（${renderQuota(teamData.owner_quota, 2)}）`);
                return;
            }
        }

        setLoading(true);
        try {
            const response = await API.post(`/api/team/${teamId}/allocate`, {
                amount: unlimited ? 0 : parseFloat(amount),
                unlimited: unlimited
            });

            if (response.data.success) {
                showSuccess(response.data.message || '消费上限设置成功');
                handleClose();
                if (onSuccess) onSuccess();
            } else {
                showError(response.data.message || '设置失败');
            }
        } catch (error) {
            console.error('设置额度失败:', error);
            showError('设置额度失败');
        } finally {
            setLoading(false);
        }
    };

    const handleClose = () => {
        if (!loading) {
            setAmount('');
            setUnlimited(false);
            onOpenChange(false);
        }
    };

    return (
        <Dialog open={open} onOpenChange={onOpenChange}>
            <DialogContent className="sm:max-w-[500px]">
                <DialogHeader>
                    <DialogTitle className="flex items-center gap-2">
                        <Wallet className="w-5 h-5" />
                        分配团队额度
                    </DialogTitle>
                    <DialogDescription>设置团队的消费上限，不会从您的个人余额扣除</DialogDescription>
                </DialogHeader>

                <form onSubmit={handleSubmit}>
                    <div className="space-y-4 py-4">
                        {/* 无限额度开关 */}
                        <div className="flex items-center justify-between space-x-2 p-4 border rounded-lg">
                            <div className="flex-1">
                                <Label htmlFor="unlimited" className="text-base font-medium cursor-pointer">
                                    无限制额度
                                </Label>
                                <p className="text-sm text-muted-foreground mt-1">启用后，团队额度上限为您的个人余额</p>
                            </div>
                            <Switch id="unlimited" checked={unlimited} onCheckedChange={setUnlimited} disabled={loading} />
                        </div>

                        {/* 金额输入 */}
                        {!unlimited && (
                            <div className="space-y-2">
                                <Label htmlFor="amount">
                                    消费上限（美元） <span className="text-destructive">*</span>
                                </Label>
                                <Input
                                    id="amount"
                                    type="number"
                                    step="0.01"
                                    min="0"
                                    placeholder="请输入团队消费上限"
                                    value={amount}
                                    onChange={(e) => setAmount(e.target.value)}
                                    disabled={loading}
                                />
                                {teamData?.owner_quota && (
                                    <p className="text-xs text-muted-foreground">您的个人余额：{renderQuota(teamData.owner_quota, 2)}</p>
                                )}
                            </div>
                        )}

                        {/* 提示信息 */}
                        <Alert>
                            <AlertCircle className="h-4 w-4" />
                            <AlertDescription>
                                {unlimited ? (
                                    <>
                                        启用无限额度时，团队的实际可用额度上限为您的个人余额
                                        {teamData?.owner_quota && (
                                            <span className="font-medium">（{renderQuota(teamData.owner_quota, 2)}）</span>
                                        )}
                                        。
                                    </>
                                ) : (
                                    '设置团队消费上限不会从您的个人余额中扣除，仅用于限制团队的总消费额度。'
                                )}
                            </AlertDescription>
                        </Alert>

                        {/* 当前额度信息 */}
                        {teamData && (
                            <div className="p-4 bg-muted rounded-lg space-y-2 text-sm">
                                <div className="flex justify-between">
                                    <span className="text-muted-foreground">当前额度：</span>
                                    <span className="font-medium">
                                        {teamData.unlimited_quota ? '无限制' : renderQuota(teamData.quota || 0, 2)}
                                    </span>
                                </div>
                                <div className="flex justify-between">
                                    <span className="text-muted-foreground">已使用：</span>
                                    <span className="font-medium">{renderQuota(teamData.used_quota || 0, 2)}</span>
                                </div>
                            </div>
                        )}
                    </div>

                    <DialogFooter>
                        <Button type="button" variant="outline" onClick={handleClose} disabled={loading}>
                            取消
                        </Button>
                        <Button type="submit" disabled={loading || (!unlimited && !amount)}>
                            {loading ? '设置中...' : '确认设置'}
                        </Button>
                    </DialogFooter>
                </form>
            </DialogContent>
        </Dialog>
    );
};

AllocateQuotaDialog.propTypes = {
    open: PropTypes.bool.isRequired,
    onOpenChange: PropTypes.func.isRequired,
    teamId: PropTypes.number.isRequired,
    teamData: PropTypes.shape({
        quota: PropTypes.number,
        used_quota: PropTypes.number,
        unlimited_quota: PropTypes.bool,
        owner_quota: PropTypes.number
    }),
    onSuccess: PropTypes.func
};

export default AllocateQuotaDialog;
