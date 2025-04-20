import React, { useState } from 'react';
import { Grid, Card, Box, Text, Avatar, Flex } from '@radix-ui/themes';
import { skinlines, skinlineSkins, asset } from '../data/data';
import { rarity } from '../data/data';
import { SkinLinesGrid } from './SkinLinesGrid';

export function SkinLinesIndex({ onSkinLineSelect }) {
  const [isLoadingIndex, setIsLoadingIndex] = useState(true);

  const handleImageLoad = () => {
    setIsLoadingIndex(false);
  };

  const getInitials = (name) => {
    return name
      .split(' ')
      .map(word => word[0])
      .join('')
      .toUpperCase()
      .slice(0, 2);
  };

  return (
    <div className='mt-[15px]'>
      <Grid className="grid grid-cols-[repeat(auto-fit,80px)] gap-3 justify-center" gap="3" width="100%" height="100%">
        {skinlines.map((skinline) => (
          <Card variant="ghost" key={skinline.id}>
            <div
              className="decoration-inherit text-ellipsis overflow-hidden text-nowrap cursor-pointer"
              title={skinline.name}
              onClick={() => onSkinLineSelect(skinline.id)}
            >
              <Box className=' h-[80px] flex items-center justify-center'>
                <Avatar
                  size="6"
                  radius="full"
                  fallback={getInitials(skinline.name)}
                />
              </Box>
              <div className="text-center mt-1 text-ellipsis overflow-hidden text-nowrap">
                <Text size="2" weight="bold">{skinline.name}</Text>
              </div>
            </div>
          </Card>
        ))}
      </Grid>
    </div>
  );
}

export function SkinLine({ skinLineId, onBack }) {
    const skinline = skinlines.find((s) => s.id === skinLineId);
    
    if (!skinline) return null;
    
    const skins = skinlineSkins(skinline.id);
  
    return (
      <div className='h-max relative mt-5'>
        <Box className='w-full'>
          <Flex display="inline-flex" direction="row" justify="center" align="center" className='items-center w-full relative mb-5'>
            <Avatar
              size="2"
              radius="full"
              fallback={skinline.name.charAt(0).toUpperCase()}
              className="mr-2"
            />
            <Text size="6" weight="bold">
              {skinline.name} Collection
            </Text>
          </Flex>
          <SkinLinesGrid skins={skins} asset={asset} rarity={rarity} />
        </Box>
      </div>
    );
  }

export default { SkinLinesIndex, SkinLine };