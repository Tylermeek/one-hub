import { useState } from 'react';
import { API } from 'utils/api';
import { showError, copy } from 'utils/common';
import { useTranslation } from 'react-i18next';
import { Card, CardContent } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import inviteImage from 'assets/images/invite/cwok_casual_19.webp';

const InviteCard = () => {
  const { t } = useTranslation();
  const [inviteUrl, setInviteUrl] = useState('');

  const handleInviteUrl = async () => {
    if (inviteUrl) {
      copy(inviteUrl, t('inviteCard.inviteUrlLabel'));
      return;
    }

    try {
      const res = await API.get('/api/user/aff');
      const { success, message, data } = res.data;
      if (success) {
        let link = `${window.location.origin}/register?aff=${data}`;
        setInviteUrl(link);
        copy(link, t('inviteCard.inviteUrlLabel'));
      } else {
        showError(message);
      }
    } catch (error) {
      return;
    }
  };

  return (
    <Card className="h-auto bg-linear-to-br from-primary/10 to-primary/5">
      <CardContent className="p-6">
        <div className="flex items-start gap-4">
          <div className="flex-1 space-y-4">
            <div>
              <h3 className="text-xl font-bold mb-2">{t('inviteCard.inviteReward')}</h3>
              <p className="text-sm text-muted-foreground">{t('inviteCard.inviteDescription')}</p>
            </div>
            <div className="flex gap-2">
              <Input value={inviteUrl} placeholder={t('inviteCard.generateInvite')} disabled className="flex-1" />
              <Button onClick={handleInviteUrl}>{inviteUrl ? t('inviteCard.copyButton.copy') : t('inviteCard.copyButton.generate')}</Button>
            </div>
          </div>
          <img src={inviteImage} alt="invite" className="w-24 h-24 opacity-80" />
        </div>
      </CardContent>
    </Card>
  );
};

export default InviteCard;
