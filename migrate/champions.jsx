import React, { useEffect, useState, memo } from 'react';
import { Grid, Text, Card, Box, Flex, Heading, Skeleton, Container, Separator, Blockquote, Section, Dialog } from '@radix-ui/themes';
import { asset, champions, _ready, championSkins, rarity, _champData, style, checkLegacy, _champChromas, getChromasForSkin, _champDataName } from '../data/data';
import { navigate, useEscapeTo, useTitle } from '../data/hooks';
import { Link, generatePath, useParams, useNavigate } from "react-router-dom";
import ChampionDialog from '../components/ChampionDialog';
import { SkinsGrid } from '../components/SkinsGrid';
import { ArrowLeft} from 'lucide-react';



export const EnhancedImage = memo(({ src, alt, className, style, onLoad, onError }) => {
  const [isLoaded, setIsLoaded] = useState(false);
  const [hasError, setHasError] = useState(false);
  const imgRef = useRef(null);
  
  useEffect(() => {
    // Reset states when src changes
    setIsLoaded(false);
    setHasError(false);
    
    // Preload image
    const img = new Image();
    img.src = src;
    img.onload = () => {
      setIsLoaded(true);
      if (onLoad) onLoad();
    };
    img.onerror = () => {
      setHasError(true);
      if (onError) onError();
    };
    
    return () => {
      img.onload = null;
      img.onerror = null;
    };
  }, [src]);
  
  return (
    <>
      {!isLoaded && !hasError && (
        <Skeleton className={className} style={style} />
      )}
      <img
        ref={imgRef}
        src={src}
        alt={alt}
        className={className}
        style={{
          ...style,
          display: isLoaded ? 'block' : 'none'
        }}
        onLoad={() => {
          setIsLoaded(true);
          if (onLoad) onLoad();
        }}
        onError={() => {
          setHasError(true);
          if (onError) onError();
        }}
      />
      {hasError && (
        <div 
          className={className} 
          style={{
            ...style,
            backgroundColor: '#1e293b',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center'
          }}
        >
          <Text size="1" color="gray">Error</Text>
        </div>
      )}
    </>
  );
});


export function ChampionsIndex({ onChampionSelect }) {
  const [isLoading, setIsLoading] = useState(true);
  const [isLoadingIndex, setIsLoadingIndex] = useState(true);

  const handleImageLoad = () => {
    setIsLoadingIndex(false);
  };
  useEffect(() => {
    _ready.then(() => setIsLoading(false));
  }, []);

  if (isLoading) {

    return (
      <Grid
        className="grid grid-cols-[repeat(auto-fit,80px)] gap-3 justify-center py-[15px]"
        gap="3"
        width="100%"
      >
        {Array.from({ length: 100 }).map((_, index) => (
          <Card variant="ghost" key={index} className='w-[120px]'>
            <Skeleton>
              <div className="w-[80px] h-[80px] bg-primary bg-contain align-middle justify-center "></div>
            </Skeleton>
            <Skeleton align="center">
              <div className="text-center mt-1 w-[60px] h-[10px] text-ellipsis overflow-hidden justify-center align-middle items-center">aaaaaaa</div>
            </Skeleton>

          </Card>
        ))}
      </Grid>
    );
  }

  return (
    <div className='mt-[15px]' >
      {/* <img src='https://skinhunter.s-ul.eu/NJCK7P1q' className='fixed object-cover '></img> */}
      <Grid className="grid grid-cols-[repeat(auto-fit,80px)] gap-3 justify-center" gap="3" width="100%" height="100%">
        {champions.map((c) => (
          <Card variant="ghost" key={c.id}>
            <div
              className="decoration-inherit text-ellipsis overflow-hidden text-nowrap cursor-pointer w-[80px] h-[110px] "
              title={c.name}
              onClick={() => onChampionSelect(c.key)}
            >
              <Box style={{ display: isLoadingIndex ? "none" : "block", width: "80px", height: "80px" }} className='bg-primary'>
                <img
                  src={`${asset(c.squarePortraitPath)}?w=80&h=80&format=webp`}
                  alt={c.name}
                  className="w-[80px] h-[80px] object-contain align-middle justify-center" 
                  style={{aspectRatio: "1 / 1"}}
                  onError={(e) => {
                    e.target.src = "/default-placeholder.png";
                  }}
                  onLoad={handleImageLoad}
                />
              </Box>
              <div className="text-center mt-1 text-ellipsis overflow-hidden text-nowrap">
                <Text size="2" weight="bold">{c.name}</Text>
              </div>
            </div>
          </Card>
        ))}
      </Grid>
    </div>
  );
}


export const SkinItem = memo(({ skin, champKey, asset, rarity }) => {
  const [isLoading, setIsLoading] = useState(true);

  const chromas = getChromasForSkin(skin);

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
            {checkLegacy(skin) && (
              <div className="absolute top-2 left-2 z-10">
                <img
                  width="24"
                  height="24"
                  src={checkLegacy(skin)[0]}
                  title={checkLegacy(skin)[1]}
                  alt="Legacy Icon"
                  style={{ filter: "brightness(0.8)" }}
                />
              </div>
            )}

            {chromas && (
              <div className="absolute top-2 right-2 z-10">
                <img
                  width="24"
                  height="24"
                  src="https://raw.communitydragon.org/latest/plugins/rcp-fe-lol-static-assets/global/default/images/skin-viewer/icon-chroma-default.png"
                  title="This skin has Chromas"
                  alt="Chromas Icon"
                  style={{
                    filter: "brightness(0.8)",
                    borderRadius: "50%",
                    cursor: "pointer",
                  }}
                />
              </div>
            )}

            <img
              src={asset(skin.tilePath)}
              alt={skin.name}
              style={{
                cursor: "url('https://cur.cursors-4u.net/games/gam-14/gam1339.cur')",
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

export function Champion({ champion, onBack }) {
  const champ = champions.find((c) => c.key === champion);
  const [champData, setChampData] = useState(null);
  const [champChromas, setChampChromas] = useState([]);
  const styles = {

    kASGrey: ["continuum_icon_attackspeed_grey.png", "Attack"],
    kASHighlight: ["continuum_icon_attackspeed.png", "Attack"],
    kAPGrey: ["continuum_icon_abilitypower_grey.png", "Magic"],
    kMagicHighlight: ["continuum_icon_abilitypower.png", "Magic"],

  }

  const roleColors = {
    fighter: "red",
    tank: "blue",
    mage: "purple",
    assassin: "orange",
    marksman: "yellow",
    support: "green",
  };

  const damageColors = {
    kPhysical: "red",
    kMixed: "yellow",
    kMagic: "purple",
  };
  useTitle(champ?.name);
  useEscapeTo("/");

  if (!champ) return navigate("/");

  const skins = championSkins(champ.id);

  useEffect(() => {
    let isMounted = true;
    _champData(champ.id).then((data) => {
      if (isMounted && data) {
        const defaultSkins = data.skins.filter((skin) => skin.isBase);
        setChampData({
          ...data,
          defaultSkins,
        });
      }
    });
    return () => {
      setChampData(null); // Limpia datos al desmontar
    };
  }, [champ.id]);

  useEffect(() => {
    let isMounted = true;
    if (!champ.id) return;
    _champDataName(champ.id).then((champData) => {
      if (!isMounted || !champData || !champData.skins) return;
      const allChromas = champData.skins
        .map(getChromasForSkin)
        .filter(Boolean)
        .flat();
      setChampChromas(allChromas);
    });
    return () => {
      setChampChromas([]); // Limpia chromas al desmontar
    };
  }, [champ.id]);

  const capitalize = (str) => {
    if (!str) return '';
    return str.charAt(0).toUpperCase() + str.slice(1);
  };

  // debugLegacySkins(skins)
  // console.log(champChromas)
  return (

      <div className='h-max  relative'>


        <Box className='w-full'>

          <Container size="2" m="8">

            {champData ? (
              <>


                <Flex direction="row" align="start" gap="4">
                  { }



                  <Box >
                    { }
                    <Flex direction="row" align="center" gap="4">
                      { }
                      <img
                        src={`${asset(champData.squarePortraitPath)}?w=80&h=80&format=webp`}
                        alt={champData.name}
                        className="w-[80px] h-[80px] object-contain"
                        style={{ aspectRatio: "1 / 1" }}
                      />

                      { }
                      <Flex direction="column" justify="center">
                        { }
                        <Heading as="h2" size="lg" mb="1" textAlign="left">
                          {champData.name}
                        </Heading>

                        { }
                        <Text as="h3" fontSize="md" color="gray.600" textAlign="left">
                          {capitalize(champData.title)}
                          <Separator my="3" size="4" />


                        </Text>
                      </Flex>
                    </Flex>
                    { }
                    <div>
                      <Blockquote fontSize="sm" textAlign="left" color="#1e1e1e" mt="3" >
                        {champData.shortBio}
                      </Blockquote>
                    </div>
                    <ChampionDialog champData={champData} roleColors={roleColors} damageColors={damageColors} style={style} asset={asset}  ></ChampionDialog>

                  </Box>

                </Flex>
              </>




            ) : (
              <Skeleton><Container size="2"><Box className='w-[full] h-[200px] bg-primary'></Box></Container></Skeleton>
            )}




          </Container>
          <Flex display="inline-flex" direction="row" justify="center" align="center" className='items-center w-full relative mb-5'>
            <img style={{ marginRight: "10px" }}
              width="35" height="auto"
              src="https://raw.communitydragon.org/latest/plugins/rcp-fe-lol-static-assets/global/default/images/skins_rewards.svg"
              className="iconimg"
            />
            <Heading as='h2' size="6" align="center" className='relative'>

              {champ.name} Skins
            </Heading>
          </Flex>

          <SkinsGrid skins={skins} asset={asset} rarity={rarity} chromas={champChromas ?? []}></SkinsGrid>
        </Box>

      </div>
    // </Section>
  );
}

