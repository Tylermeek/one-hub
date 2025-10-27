import { useSelector } from 'react-redux';
import { useState, useCallback } from 'react';
import { API } from 'utils/api';
import { replaceChatPlaceholders, getChatLinks } from 'utils/common';
import { MessageSquare, Monitor } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Button } from '@/components/ui/button';

const QuickStartCard = () => {
  const { t } = useTranslation();
  const [key, setKey] = useState('');
  const siteInfo = useSelector((state) => state.siteInfo);
  const chatLinks = getChatLinks(false);
  const baseServer = siteInfo.server_address;

  const getProcessedUrl = useCallback(
    (url, key) => {
      let server = baseServer || window.location.host;
      server = encodeURIComponent(server);
      const useKey = 'sk-' + key;
      return replaceChatPlaceholders(url, useKey, server);
    },
    [baseServer]
  );

  const handleClick = async (url) => {
    if (!key) {
      try {
        const res = await API.get(`/api/token/playground`);
        const { success, message, data } = res.data;
        if (success) {
          setKey(data);
          window.open(getProcessedUrl(url, data), '_blank');
        } else {
          console.log('message', message);
        }
      } catch (error) {
        console.error('Failed to get token:', error);
      }
    } else {
      window.open(getProcessedUrl(url, key), '_blank');
    }
  };

  // 分离不同类型的链接
  const safeChatLinks = chatLinks || [];
  const webLinks = safeChatLinks.filter((option) => option.url.startsWith('http'));
  const appLinks = safeChatLinks.filter((option) => !option.url.startsWith('http'));

  return (
    <Card className="h-auto">
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <MessageSquare className="w-5 h-5" />
          {t('dashboard_index.quickStart')}
        </CardTitle>
        <CardDescription>{t('dashboard_index.quickStartTip')}</CardDescription>
      </CardHeader>
      <CardContent className="space-y-6">
        {/* Web 应用 */}
        {webLinks.length > 0 && (
          <div className="space-y-3">
            <h4 className="text-sm font-medium text-muted-foreground">Web 应用</h4>
            <div className="grid grid-cols-2 gap-3">
              {webLinks.map((option, index) => (
                <Button key={index} variant="default" onClick={() => handleClick(option.url)} className="w-full">
                  <MessageSquare className="w-4 h-4 mr-2" />
                  {option.name}
                </Button>
              ))}
            </div>
          </div>
        )}

        {/* 桌面应用 */}
        {appLinks.length > 0 && (
          <div className="space-y-3">
            <h4 className="text-sm font-medium text-muted-foreground">桌面应用</h4>
            <div className="grid grid-cols-2 gap-3">
              {appLinks.map((option, index) => (
                <Button key={index} variant="outline" onClick={() => handleClick(option.url)} className="w-full">
                  <Monitor className="w-4 h-4 mr-2" />
                  {option.name}
                </Button>
              ))}
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  );
};

export default QuickStartCard;
