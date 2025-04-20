import React, { useState, memo } from 'react';
import {
  Card,
  Box,
  Dialog,
  Flex,
  Button,
  Skeleton,
  ScrollArea,
  Callout,
  Tabs,
  RadioCards,
} from '@radix-ui/themes';
import { InfoCircledIcon, DownloadIcon } from '@radix-ui/react-icons';
import { KhadaUrl } from '../data/data';
import { useUser } from '../context/usercontext';
import { useDownloadSkin } from '../context/download';
import { motion, AnimatePresence } from 'framer-motion';
import SkinDialWrapper from './SkinDialWrapper';
import { SkinItem } from '../pages/champions';

const TabAnimation = {
  initial: { opacity: 0, y: 20 },
  animate: { opacity: 1, y: 0 },
  exit: { opacity: 0, y: -20 },
  transition: { duration: 0.3 }
};
// SkinLineItem component for individual skin items in the grid
export const SkinLineItem = memo(({ skin, asset, rarity }) => {
  const [isLoading, setIsLoading] = useState(true);

  const handleImageLoad = () => {
    setIsLoading(false);
  };

  return (
    <Box
      style={{ willChange: "transform" }}
      className="relative rounded-lg overflow-hidden hover:scale-105 transition-transform"
      key={skin.id}
    >
      <Flex
        align="center"
        direction="column"
        justify="center"
        className="items-center"
      >
        {isLoading && (
          <Skeleton>
            <Box style={{ width: "200px", height: "200px" }} />
          </Skeleton>
        )}

        <Box
          style={{
            display: isLoading ? "none" : "block",
            width: "200px",
            height: "200px",
          }}
          className="bg-primary relative"
        >
          {skin.isLegacy && (
            <div className="absolute top-2 left-2 z-10">
              <img
                width="24"
                height="24"
                src="https://raw.communitydragon.org/latest/plugins/rcp-fe-lol-static-assets/global/default/images/summoner-icon/icon-legacy.png"
                title="Legacy Skin"
                alt="Legacy Icon"
                style={{ filter: "brightness(0.8)" }}
              />
            </div>
          )}

          <img
            src={asset(skin.tilePath)}
            alt={skin.name}
            style={{
              cursor: "pointer",
              objectFit: "cover",
              width: "100%",
              height: "100%"
            }}
            onLoad={handleImageLoad}
          />
        </Box>

        <Flex 
          direction="row" 
          align="center" 
          className="w-[200px] px-2 py-1 bg-black/50 backdrop-blur-sm"
        >
          {rarity(skin) && (
            <img
              className="mr-2"
              width="16"
              height="16"
              src={rarity(skin)[0]}
              title={rarity(skin)[1]}
              alt={rarity(skin)[1]}
            />
          )}
          <span className="text-sm text-white truncate">
            {skin.name}
          </span>
        </Flex>
      </Flex>
    </Box>
  );
});

// SkinLineDialog component for the detailed view
export const SkinLineDialog = ({ skin, asset, rarity, chromas }) => {
  const [isLoadingDs, setIsLoadingDs] = useState(true);
  const [isInstalled, setIsInstalled] = useState(false);
  const [selectedChroma, setSelectedChroma] = useState(null);
  const { userData } = useUser();
  const downloadSkin = useDownloadSkin();
  const filteredChromas = chromas?.filter(chroma => chroma.origin === skin.id) || [];
  
  const handleImageLoads = () => setIsLoadingDs(false);

  
  
  const handleSelectChroma = (value) => {
    setSelectedChroma(prev => prev === value ? null : value);
  };

  const handleDownload = async () => {
    const idToDownload = selectedChroma ? parseInt(selectedChroma) : skin.id;
    const selectedChromaData = selectedChroma
      ? chromas.find(chroma => String(chroma.id) === selectedChroma)
      : null;
      
    await downloadSkin(
      Math.floor(skin.id / 1000),
      idToDownload,
      setUserData,
      skin,
      selectedChromaData
    );
  };

  return (
    <Dialog.Content className="custom-dialog" maxWidth="850px">
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        exit={{ opacity: 0, y: -20 }}
      >
        <Dialog.Title>
          <Flex direction="row" align="center" mt="2">
            {rarity(skin) && (
              <motion.img
                initial={{ scale: 0.8 }}
                animate={{ scale: 1 }}
                className="mr-2"
                width="20"
                height="auto"
                src={rarity(skin)[0]}
                title={rarity(skin)[1]}
                alt={rarity(skin)[1]}
              />
            )}
            {skin.name}
          </Flex>
        </Dialog.Title>

        {/* Main content structure remains similar but with animations */}
        <Flex direction={{ md: "row", xl: "row" }} gap="4" className="w-full">
          {/* Left side - Skin image and description */}
          <Box style={{ flex: 1 }}>
            <motion.div {...TabAnimation}>
              <Card size="4" style={{ width: "100%", maxWidth: "450px" }}>
                {/* Image and description content */}
              </Card>
            </motion.div>
          </Box>

          {/* Right side - Chromas and controls */}
          <Flex direction="column" gap="3" style={{ flex: 1 }}>
            <Tabs.Root defaultValue="circles">
              <Tabs.List>
                <Tabs.Trigger value="circles">Circles</Tabs.Trigger>
                <Tabs.Trigger value="images">Images</Tabs.Trigger>
              </Tabs.List>

              <AnimatePresence mode="wait">
                <Tabs.Content value="circles">
                  <motion.div {...TabAnimation}>
                    {/* Circles tab content */}
                    <ScrollArea type="auto" className="max-h-64">
                      <RadioCards.Root value={selectedChroma} onValueChange={handleSelectChroma}>
                        {/* Chroma options */}
                      </RadioCards.Root>
                    </ScrollArea>
                  </motion.div>
                </Tabs.Content>

                <Tabs.Content value="images">
                  <motion.div {...TabAnimation}>
                    {/* Images tab content */}
                    <ScrollArea type="auto" className="max-h-64">
                      <RadioCards.Root value={selectedChroma} onValueChange={handleSelectChroma}>
                        {/* Image options */}
                      </RadioCards.Root>
                    </ScrollArea>
                  </motion.div>
                </Tabs.Content>
              </AnimatePresence>
            </Tabs.Root>

            {/* Credits callout */}
            <motion.div
              initial={{ opacity: 0, y: 10 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: 0.2 }}
            >
              <Callout.Root variant="surface">
                <Callout.Icon><InfoCircledIcon /></Callout.Icon>
                <Callout.Text>
                  This is going to consume a credit {userData?.fichasporskin > 0 && ` (${userData.fichasporskin} credits left)`}
                </Callout.Text>
              </Callout.Root>
            </motion.div>
          </Flex>
        </Flex>

        {/* Dialog actions */}
        <Flex className="justify-between mt-4">
          <Dialog.Close asChild>
            <Button size="3" variant="soft" color="gray">Close</Button>
          </Dialog.Close>
          
          <motion.div whileHover={{ scale: 1.05 }} whileTap={{ scale: 0.95 }}>
            <Button
              onClick={handleDownload}
              disabled={!userData || userData.fichasporskin <= 0}
              size="3"
              color="gray"
            >
              <DownloadIcon />
              Download {selectedChroma ? 'Chroma' : 'Skin'}
            </Button>
          </motion.div>
        </Flex>
      </motion.div>
    </Dialog.Content>
  );
};

export const SkinLinesGrid = ({ skins, asset, rarity }) => {
  const getChampKey = (skinId) => {
    const champId = Math.floor(skinId / 1000);
    return `champion-${champId}`; // Formato básico para el champKey
  };

  return (
    <Box className="grid grid-cols-[repeat(auto-fit,250px)] gap-3 justify-center items-center pb-7 justify-items-center">
      {skins.map((skin) => (
        <Dialog.Root key={skin.id}>
          <Dialog.Trigger asChild>
            <div className="cursor-pointer w-fit h-fit">
              {/* Usar SkinItem en lugar de SkinLineItem */}
              <SkinItem
                skin={skin}
                champKey={getChampKey(skin.id)}
                asset={asset}
                rarity={rarity}
              />
            </div>
          </Dialog.Trigger>
          <SkinDialWrapper
            skin={skin}
            asset={asset}
            rarity={rarity}
          />
        </Dialog.Root>
      ))}
    </Box>
  );
};

export default SkinLinesGrid;