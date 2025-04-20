import React, { useState, useEffect } from 'react';
import { _champData, _champDataName, getChromasForSkin } from '../data/data';
import SkinDial from './SkinDial'; // Asegúrate de importar SkinDial correctamente
import {
    Dialog,
    Flex,
    Spinner,
  } from '@radix-ui/themes';
const SkinDialWrapper = ({ skin, asset, rarity }) => {
  // Estados para manejar chromas, champKey y estado de carga
  const [chromas, setChromas] = useState([]);
  const [isLoading, setIsLoading] = useState(true);
  const [champKey, setChampKey] = useState(null);

  useEffect(() => {
    const fetchData = async () => {
      setIsLoading(true);
      // Obtener el ID del campeón a partir del ID de la skin
      // Suponemos que skin.id tiene el formato champId * 1000 + skinId
      const champId = Math.floor(skin.id / 1000);
      try {
        const champData = await _champData(champId);
        if (champData) {
          setChampKey(champData.key); // Establecer champKey
          const additionalData = await _champDataName(champId);
          if (additionalData && additionalData.skins) {
            // Obtener todos los chromas para las skins del campeón
            const allChromas = additionalData.skins
              .map(getChromasForSkin)
              .filter(Boolean)
              .flat();
            setChromas(allChromas);
          }
        }
      } catch (error) {
        console.error("Error fetching champion data:", error);
      } finally {
        setIsLoading(false);
      }
    };
    fetchData();
  }, [skin]); // Volver a ejecutar si cambia la skin

  // Mostrar un indicador de carga mientras se obtienen los datos
  if (isLoading) {
    return (
      <Dialog.Content maxWidth="120px" maxHeight="80px">
        <Flex direction="column" align="center" justify="center" gap="4">
          <Spinner />
        </Flex>
      </Dialog.Content>
    );
  }

  // Renderizar SkinDial con los datos obtenidos
  return (
    <SkinDial
      skin={skin}
      champKey={champKey}
      asset={asset}
      rarity={rarity}
      chromas={chromas}
    />
  );
};

export default SkinDialWrapper;