<?php

/* CLASE PARA CONTENER DATOS */

class ContDatos
{
    /*
      var $imprime;   // si se permite la impresion
      // Datos que contiene
      var $nombreTabla;
      var $TablaBase;   // Nombre de la tabla base
      var $tablas;    // Array de Tablas contenidas
      var $TablaTemporal; // Objeto TablaTemporal

      var $AsientoTemp; // Array de Asiento
      var $Minuta;    // Objeto minuta
      var $registrosEx;   // Array con Objetos RegistroEx

      var $CabeceraMov;   // Cabecera de Datos para los Ingresos (Contiene otro contenedor)

      var $CuerpoMov;   // Array de Contenedores con las Tablas que se graban
      var $preConsultas;  // Array de Contenedores con las Tablas que se consultan previamente a la principal

      // Propiedades

      var $grafico;   // Array con los graficos contenidos

      var $titulo;  // Titulo de lo que contiene

      var $xml;       // Nombre del archivo XML que se lee;
      var $xmlOrig;     // Nombre del archivo XML Original (para cabeceras)
      var $xmlpadre;    // Nombre del archivo XML Cuando es una subconsulta
      var $xmlReferente;  // Nombre del archivo XML Referente (Para ingresos que llenan tablas al volver)

      var $llenoReferente;  // al procesar  los datos lleno el contenedor referente;
      var $exportasuma;   // lleno el contenedor referente con la suma de los conceptos;

      var $obs;       // Variable que almacena Las observaciones del Contenedor
      var $tipoAbm;     // tipo de ABM donde se utilizara
      var $subtipo;     // ej: en ingresos (Vertical u Horizontal)
      var $importadatos;  // xml que tiene (en los ingresos) los origenes de los datos a importar

      var $noForm;
      var $imgpath;     // sub path para las fotos
      //var $path;      // sub path para los archivos

      var $chkdup;    // Indica si antes de hacer un insert debe efectuar un select para evitar claves duplicadas y hacer un update si encuentra

      var $grabaCab;    // Indica si se graba o no la cabecera en una tabla
      var $grabaFinal;  // Indica si graba la suma de los campos 1 sola vez

      var $sololectura;   // si los datos estan en modo solo lectura
      var $CampoCant;   // Campo de la tabla principal que determina la cantidad de registros a grabar en los ingresos de movimientos
      var $CampoCond;   // indica el campo que condicion la grabacion de los movimientos en los ingresos
      var $CampoCondExp;  // indica (PARA CUANDO ES GRABACION AL FINAL la condicion de la grabacion de los movimientos en los ingresos

      var $detalle;   // nombre(s?) del archivo xml que contiene la descripcion del detalle (para consultas con cabecera y detalle)
      var $esDetalle;   // indica si el contenedor es el detalle de otro
      var $showCab;     // indica si muestro o no la cabecera del detalle (por defecto NO)
      var $insertaABM;  // especifica si el abm tiene que mostrar el boton insertar
      var $modificaABM;   // especifica si el abm tiene que mostrar el boton Modificar

      var $retorna;     // nombre del campo que retorna cuando es un contenedor interno (ej. parmetro)

      var $tabindex;    // variable global al Contenedor con el indice de tabulacion

      // SQL
      var $Condicion;
      var $orden;     // Array de campos de orden
      var $group;     // Para los group by
      var $resultSet;
      var $Joins;
      var $tipoJoin;
      var $codigoInsert; // Codigo especial que se ejecuta en los inserts de los ingresos de movimientos

      var $cadenasSQL;    // Array de SQL que se enviaran al contenedor referente

      var $sqlExterna2; // Array que recibe las cadenas SQL del contenedor referido

      var $filtroPrincipal;
      var $campofiltroPrincipal;

      var $filtros;   //Array de Filtros Fijos (proximamente variables)
      var $autofiltro;

      var $sumaCampo;
      var $acumulaCampo;

      var $llenoTemporal;
      var $graboasiento;
      var $muestraasiento;

      var $isInner;
      var $preloadData;
      var $fileResultSet;
      var $paginar;
      var $swap;
      var $modificar;
      var $inserta;
      var $borra;
      var $rss;
      var $dirxml;
      var $barraDrag;
      var $campoRetorno;
      var $preFetch;
      var $cancel;
      var $tableBorder;
      var $detallado;
      var $paginaActual;
      var $maxshowresult;
      var $lastSearchString;
      var $form;
      var $forzado;
      var $editable;
      var $TotalRegistros;
      var $prefijoResultados;
      var $sufijoResultados;
      var $resize;
      var $__inline;
      var $ldifExport;
      var $notifica;
      var $imprimetanda;
      var $col1;
      var $col2;
      var $conBusqueda;
      var $exportData;
      var $Mnoms;
      var $style;
      var $width;
      var $ancho;
      var $subdir;
      var $innerTablaData;
      var $isTableEx;
      var $procesa;
      var $tooltip;
      var $btnconfirma;
      var $cierraproceso;
      var $grillasContenidas;

      var $validations;

      var $__inlineid;
      var $autoing;
      var $deleteRow;
      var $ordenaTemporal;
      var $lastLogcount;
      var $cerrar_proceso;
      var $limit;
      var $ldap;
      var $oldSearch;
      var $eventosXML;
      var $log;
      var $_menuId;
      var $sqlhardcoded;
      var $customSelect;
      var $replaceInto;

      var $postGrabacion;
      var $externalPrintScripts;
      var $printScripts;
      var $printContainerReaders;
      var $externalPrintContainerReaders;
      var $close;

      var $tituloAbm;

      // el PDF no lleva cabecera;
      var $PDFsincabecera;
      var $PDForientacion;

     */

    /**
     * CONSTRUCTOR
     * Por ahora el constructor toma el nombre de la tabla principal
     * Mas adelante va a ser indistinto el orden
     *
     * @param string $inTbl   Nombre de la Tabla Principal
     * @param string $tit     Titulo de la Tabla
     * @param string $tipoAbm Tipo de Abm a Crear
     */
    public function __construct($inTbl, $tit = '', $tipoAbm = '')
    {

        /* nombre de la tabla principal */

        if ($tit == '')
            $tit = $inTbl;
        $this->titulo = $tit;
        $this->tipoAbm = $tipoAbm;
        $this->tipo = $tipoAbm;
        $this->TablaBase = $inTbl;
        $this->addTabla($inTbl);
        $this->TablaTemporal = new TablaDatos($inTbl);
        $this->setInstance();
     //   $this->__row = $_REQUEST['__row'];
     //   $this->setInstance($this->__row);

    }


    public function destroy()
    {
      
        //loger('xml: '.$this->xml, 'session.log');      

        // destroy inner containers
        foreach ($this->tablas[$this->TablaBase]->campos as $fieldName => $Field) {

            if (isset($Field->contExterno)) {
              $Field->contExterno->destroy();
            }

            if ($Field->getContenedorAyuda()){
              $Field->getContenedorAyuda()->destroy();
            }

        }
        // delete Pre Query
        if (is_array($this->preConsultas))
        foreach ($this->preConsultas as $key => $preCont) {
            $preCont->destroy();
        }


        // relations for processing
        if (is_array($this->CuerpoMov))
        foreach ($this->CuerpoMov as $key => $CuerpoMov) {
          if ( method_exists ( $CuerpoMov, 'destroy'))
            $CuerpoMov->destroy();
            
        }      
        if (is_array($this->CabeceraMov))
        foreach ($this->CabeceraMov as $key => $CabeceraMov) {
            $CabeceraMov->destroy();
            
        }      

        $instance = $this->getInstance();
        
        unset($_SESSION['xml'][$instance] );  
        unset($_SESSION['instances'][$instance]);

    }

    /**
      Magic method
     */
    public function __toString()
    {
       return get_class($this);
         //$this->_instance = UID::getUID($this->TablaBase.'_'.$row.'_');

     }


    /**
      Set Programm Instance, to be used in logging
     */
    public function setInstance($row='')
    {
       $this->_instance = UID::getUID($this->TablaBase);
         //$this->_instance = UID::getUID($this->TablaBase.'_'.$row.'_');

     }

     /**
      Set Programm Instance, to be used in logging
     */
    public function getInstance()
    {
         return $this->_instance;
    }

     /**
     * Get parameter value to field
     */

    public function getParametros($campo)
    {
        if (isset($campo->ContParametro)) {
            $campo->ContParametro->cargoCampos();
            $valor = $campo->ContParametro->getCampo($campo->ContParametro->retorna)->getValor();
            $campo->setValor($valor);

            return true;
        }

        return false;
    }

     /**
     * Update parameter Value
     */

    public function setParametros($campo, $valor, $retorna = false)
    {
        if (isset($campo->ContParametro)) {
            $campo->ContParametro->getCampo($campo->ContParametro->retorna)->setNuevoValor($valor);
            if ($retorna) {
                   return $campo->ContParametro->getUpdate();
             }
            $campo->ContParametro->Update(null, $retorna);
        }
    }

    /**
     * Grabacion de los Registros UNICOS que toman los datos de la cabecera y/o
     * de la suma de los campos y NO del valor de la tabla temporal
     * @param ContDatos $relationContainer       Contenedor de generacion de Movimientos
     * @param strign    $numeroMinuta Numero de Minuta Contabel
     */
    public function GrabarRegistrosUnicos(&$relationContainer, $numeroMinuta = '')
    {
        if ($this->xmlpadre != '' || isset($this->parentInstance)) {
            $contPadre = new ContDatos("");
            // Tratar de reemplazar esta DES-SERIALIZACION (TARDA MUCHO)

            $contPadre = Histrix_XmlReader::unserializeContainer(null, $this->parentInstance);
        }

        // Asigno el valor de cada Campo tomandolo de la Tabla temporal
      if ($relationContainer->tablas[$relationContainer->TablaBase]->campos != '') 
        foreach ($relationContainer->tablas[$relationContainer->TablaBase]->campos as $Ncampo => $campoMovim) {

            // Obtengo el nombre del campo que se utiliza en la asignacion

            $nomCampoTemp  = $campoMovim->campoTemp;
            $nomCampoCab   = $campoMovim->campoCab;
            $nomCampoPadre = $campoMovim->campoPadre;

            /*
             * probar la cobranza de restiffo ANTES DE ACTIVAr
             *
             *
             * if (isset($campoMovim->valor)) {
                  $relationContainer->setNuevoValorCampo($campoMovim->NombreCampo, $campoMovim->valor);
                  $relationContainer->setCampo($campoMovim->NombreCampo, $campoMovim->valor);
              }
             */
              /**
             * Cuando es una consulta anidada busco los datos del contenedor padre
             */

            $relationContainer->CargoTablaTemporalDesdeCampos(true);
            //$arrayTT = $relationContainer->TablaTemporal->datos();
            $arrayTT = $relationContainer->TablaTemporal->Tabla;


            if ($campoMovim->valor != '') {
              $relationContainer->setFieldValue($campoMovim->NombreCampo, $campoMovim->valor, 'both');
            }


            if ($nomCampoPadre != '') {
                if (is_object($contPadre)) {

                    $campoPadre = $contPadre->getCampo($nomCampoPadre);
                    $contPadre->getParametros($campoPadre);

                    $valorPadre = $campoPadre->valor;

                    if (count($campoPadre->opcion) > 0 && $campoPadre->TipoDato != "up" && $campoPadre->valoropcion == 'true') {
                        $valorPadre = $campoPadre->opcion[$valorPadre];
                        if (is_array($valorPadre))
                          $valorPadre = current($valorPadre);
                    }

                    // Tomo el valor de la cabecera y lo cargo en el registro del Movimiento
                    // Si coninciden las relaciones
                    // Formateo el campo
                    
                    $formatoCampo = $campoMovim->Formato;

                    // FORMATEO EXPLICITO
                    
                    if ($formatoCampo != '')
                      $valorPadre = sprintf($formatoCampo, $valorPadre);

                    if ($campoMovim->valor != '')
                      $valorPadre = $campoMovim->valor;

                    $relationContainer->setFieldValue($campoMovim->NombreCampo, $valorPadre, 'both');
//                  loger('Field!!!!!: '.$nomCampoPadre.' missing'.$valorPadre, 'process.log');

                    $valor = $valorPadre;
                    $campoMovim->updateSetters($valorPadre);

                    $arrayTT[0][$campoMovim->NombreCampo] = $valor;

                } else {
                  loger('Field: '.$nomCampoPadre.' missing', 'process.log');
                }
              }

              if (isset($this->CabeceraMov))
                foreach ($this->CabeceraMov as $nCabecera => $containerCabecera) {
                    $detallecampo = '';
                    // tomo el valor del numerador aca de nuevo

                    foreach ($containerCabecera->tablas[$containerCabecera->TablaBase]->campos as $numCamCab => $campoCabecera) {
                      if ($campoCabecera->NombreCampo == $nomCampoCab) {
                        $containerCabecera->getParametros($campoCabecera);

                        $valorCabecera = $campoCabecera->valor;

                        if (count($campoCabecera->opcion) > 0 && $campoCabecera->TipoDato != "check" && $campoCabecera->valoropcion == 'true') {

                            $valorCabecera = $campoCabecera->opcion[$valorCabecera];
                            if (is_array($valorCabecera))
                                    $valorCabecera = current($valorCabecera);
                        }

                        // Tomo el valor de la cabecera y lo cargo en el registro del Movimiento
                        // Si coninciden las relaciones
                        // Formateo el campo
                        $formatoCampo = $campoMovim->Formato;
                        /* FORMATEO EXPLICITO */
                        if ($formatoCampo != '')
                            $valorCabecera = sprintf($formatoCampo, $valorCabecera);

                        if ($campoMovim->valor != '')
                            $valorCabecera = $campoMovim->valor;

                        $relationContainer->setFieldValue($campoMovim->NombreCampo, $valorCabecera, 'both');
                        $valor = $valorCabecera;

                      } else {
                        if ($campoMovim->valor != '') {
                          $valor = $campoMovim->valor;
                          $relationContainer->setFieldValue($campoMovim->NombreCampo, $valor, 'both');
                        }
                      }

                            //vinculo detalle para la impresion

                            $valorCabecera = $campoCabecera->getValorGen($campoCabecera->valor, $campoCabecera->TipoDato);

                            $reemplazablesNombres[$campoCabecera->NombreCampo] = '[__' . $campoCabecera->NombreCampo . '__]';
                            $reemplazablesValores[$campoCabecera->NombreCampo] = addslashes($valorCabecera);

                            $detallecampo = $campoCabecera->getUrlVariableString($valorCabecera);
                            if ($detallecampo != '')
                                $det .= '&' . $detallecampo;
                        }
                    }

                if ($nomCampoTemp != '') {

                    $objCampoTempo = $this->getCampo($nomCampoTemp);
                    if (isset($objCampoTempo->suma)) {
                        $valorAcumulado = $objCampoTempo->Suma;
                    } else {
                        $valorAcumulado = $objCampoTempo->ultimo;
                    }

                    if (trim($valorAcumulado) == '') {
                        $valorAcumulado = $objCampoTempo->valor;
                    }
                    if (trim($valorAcumulado) != '') {
                        $relationContainer->setNuevoValorCampo($campoMovim->NombreCampo, $valorAcumulado);
                    }

                    $reemplazablesNombres[$nomCampoTemp] = '[__' . $nomCampoTemp . '__]';
                    $reemplazablesValores[$nomCampoTemp] = $valorAcumulado;

                }
                
                // Using Last select Id
                if ($campoMovim->lastId != '') {
                     
                    if (isset($this->lastInsertedId[$campoMovim->lastId])) {
                        $valor = $this->lastInsertedId[$campoMovim->lastId];
                    } else {
                        $valor = '[__lastId__' . $campoMovim->lastId . '__]';
                    }
                    $relationContainer->setNuevoValorCampo($campoMovim->NombreCampo, $valor);

                }

                // excepcion para la grabacion de los asientos (ver como generalizar despues)
                if ($this->graboasiento) {
                    if ($campoMovim->esNroMinuta == true) {
                        $valor = $numeroMinuta;

                        $relationContainer->setFieldValue($campoMovim->NombreCampo, $valor, 'both');

                    }
                }


                // Detalle de los Nmovim
                // Se utiliza para Imprimir desde los xml que se procesan.
                if ($relationContainer->xmlImpresion != '') {
                    
                    $detallecampoMovim = $campoMovim->getUrlVariableString($valor);
                    
        //VincularDetalle($valor);

                    if ($detallecampoMovim != '') {
                        $detNmovim .= '&' . $detallecampoMovim;
                    }

                }
                
            }

        $codigoEvaluado = '';
        if (isset($relationContainer->codigoInsert)) {

            $codigoEvaluado = $relationContainer->calculoExpresion($relationContainer->codigoInsert, $arrayTT[0]);
//            loger($codigoEvaluado, 'exe');
        }

        if ($relationContainer->tipoAbm != 'nosql') {
            if ($relationContainer->tipoAbm == 'update') {
                $cadenaSQL = $relationContainer->getUpdate(true) . ' ';
            } else {
              $cadenaSQL = $relationContainer->getInsert() . ' ';
            }

        }

        $campoCant    = $relationContainer->CampoCant;
        $campoCond    = $relationContainer->CampoCond;
        $campoCondExp = $relationContainer->CampoCondExp;

        $arraydatos = $this->TablaTemporal->Tabla;

        // Recorro el la tabla temporal
        if (isset($arraydatos) && $campoCond != '') {
            foreach ($arraydatos as $numTT => $rowTT) {

                //  loger('EVALUO campos en '.$relationContainer->xml, 'ev');
                //  loger($rowTT, 'ev');
                if (is_array($arrayTT[0])) {
                  $rowTT = array_merge($arrayTT[0], $rowTT);
                }


                if ( $rowTT[$campoCond] == 'false'
                  || $rowTT[$campoCond] == false
                  || $rowTT[$campoCond] == 0
                  || $rowTT[$campoCond] == '0'
                  || $rowTT[$campoCond] == '') {

                    $noexec = 'true';
                } else {
                    $noexec = 'false';
                }

                /*
                // rewrite validation.... test
                if ($rowTT[$campoCond] === true || $rowTT[$campoCond] === 'true') {
                    $noexec = 'false';
                }
                */

            }
        }


        if ($campoCondExp != '') {
  
//  loger('EVALUO CAMPO COND EX EN xml '.$relationContainer->xml, 'ev');
//  loger('campo: '.$campoCond.' : '.$rowTT[$campoCond].'  = '.$campoCondExp .'  '.$noexec , 'ev');

            $noexec = 'false';
            $evaluado = null;
            $objCampoCond = $this->getCampo($campoCond);
            $valorUltimo = $objCampoCond->ultimo;

            if (trim($valorUltimo) != '') {

            } else {
                if ($objCampoCond != '') {
                    $valorUltimo = $objCampoCond->valor;
                } else {
                    // busco en el contenedor padre
                    //   $campoPadre  = $contPadre->getCampo($nomCampoPadre);
                    
                    if (is_object($contPadre)) {
                      $valorUltimo = $contPadre->getCampo($campoCond)->valor;
                    }
                    

                    if (isset($this->CabeceraMov)) {
                        foreach ($this->CabeceraMov as $nCabecera => $containerCabecera) {  
                            if ($containerCabecera->getCampo($campoCond))
                              $valorUltimo = $containerCabecera->getCampo($campoCond)->valor;

                        }
                    }
                }
            }

            if ($valorUltimo == '')
                $valorUltimo = 0;

            $expresionaevaluar = '$evaluado = (' . $valorUltimo . ' ' . $campoCondExp . ')?"true":"false";';
//            loger($relationContainer->xml . ' ' . $expresionaevaluar, 'ev');
            @eval($expresionaevaluar);
            if ($evaluado == 'false') {
                $noexec = 'true';
            }
//      loger($noexec, 'ev');
        }



        // print por relationContainer
        // se ejecuta al final de cada proceso del relationContainer para
        // imprimir dentro de un proceso N veces
        // a menos que lleno el contenedor referente donde guardo el comando para ejecutarlo al final
        //
        // Implementado el metodo de impresion directa

        if ($relationContainer->xmlImpresion != ''  && $noexec != 'true') {


            $printScript = $this->printExternal($relationContainer->xmlImpresion, $detNmovim, $relationContainer->dirImpresion);


            // new nethod of direct printing
            // create Loader for print container

            $printContainer = $this->printContainer($relationContainer->xmlImpresion, $detNmovim, $relationContainer->dirImpresion);

            unset($detNmovim);

            if ($this->llenoReferente == 'true') {
                $this->printScripts[] = $printScript;

                $this->printContainerReaders[] = $printContainer;
                 echo '<script type="text/javascript">' . $printScript . '</script>';
            } else {
                if ($relationContainer->directprint != '') {

                    $printContainer->printer = $relationContainer->directprint;
                    $this->externalPrintContainerReaders[] = $printContainer;

                } else {

                    echo '<script type="text/javascript">' . $printScript . '</script>';
                }
            }

        }


        if (strpos($cadenaSQL, '[__') && strpos($cadenaSQL, '__]'))
            $cadenaSQL = trim(str_replace($reemplazablesNombres, $reemplazablesValores, $cadenaSQL));

        /**
         * Si viene de un referente entonces las posibles sentencias SQL
         * Se tienen que ejecutar cuando se confirme el referente y no el referido
         */
        if ($noexec != 'true') {
            if ($this->llenoReferente == 'true') {
                $this->cadenasSQL[$relationContainer->xml][] = $cadenaSQL;
                $this->evalStatements[] = $codigoEvaluado;

                $cadenaSQL = '';
            } else {

              try {
                eval($codigoEvaluado);
              } catch (Exception $e) {
                loger($e, 'eval_errors.log');
              }

                if (trim($cadenaSQL) != '') {
                    $updateResponse = updateSQL($cadenaSQL, $relationContainer->tipoAbm, $this->xml);

                    // rollback support
                    if ($updateResponse === -1) {
                        return -1;
                    }

                    $this->lastInsertedId[$relationContainer->xml] = $updateResponse;
                }
            }
        }
    }

    /**
     * Envia el SQL generado al contenedor referente
     * @param ContDatos $contReferente Contenedor referente
     *
     */
    public function SqlAlReferente(&$contReferente = null)
    {

      // check to fill parent Containers
        if ($this->llenoReferente == 'false' ) return;
        
        // Obtengo la tabla temporal de destino
        $xmlref = $this->xmlReferente;
        
        if ($contReferente == null) {

            // tengo el contenedor original
          $reloadParent = true;

            $contReferente = new ContDatos("");
            if ($xmlref != '') {
                $contReferente = Histrix_XmlReader::unserializeContainer(null, $this->parentInstance);
                //if ($contReferente->xml != $xmlref)
                //        $reloadParent = false;

                // if is nested inside another container we need to bubble up
                if (!isset($contReferente->CuerpoMov) && isset($contReferente->parentInstance)){
                  $contReferente = Histrix_XmlReader::unserializeContainer(null, $contReferente->parentInstance);                      
                }

            }

        }


//      loger($this->idxml.' sube a '.$contReferente->xml, 'updatessql.log');

        
        // set last instance on parent Object
        if ($this->llenoReferente) {
            $contReferente->childContainers[$this->xml] = $this->getInstance();
            //$row = ($this->__row !='' )?$this->__row:0;
            //$contReferente->childContainers[$this->xml][$row] = $this->getInstance();
          }

    
    // update Last inserted ID
        if (isset($this->lastInsertedId)) {
            if (isset($contReferente->lastInsertedId)) {
                $contReferente->lastInsertedId = array_merge($contReferente->lastInsertedId, $this->lastInsertedId);
            } else {
                //
                // $contInterno->lastInsertedId = $this->lastInsertedId;
            }
        }
        
        
        // guardo en el contenedor referente las cadenas de SQL que se tienen que ejecutar

    // copy foreing sql statements to current SQL statements
        if ($this->sqlExterna2 != '') {
            foreach ($this->sqlExterna2 as $xml => $sqlArray) {

                foreach($sqlArray as $num => $sql1){
                    $this->cadenasSQL[$xml][] = $sql1;
                }

            }
//            unset($this->sqlExterna2); // remove after use
        }


    // Propagate external print scripts
        if ($this->externalPrintScripts != '') {
            foreach ($this->externalPrintScripts as $pscr => $pscript) {
                $this->printScripts[] = $pscript;

            }
            //$this->printScripts =  array_reverse($this->printScripts);
        }

        // delete previous sql statements
        // not to  duplicate statements
        if(isset($contReferente->sqlExterna2)){
          unset($contReferente->sqlExterna2[$this->xml]);
        }

        if ($this->cadenasSQL != '') {
            foreach ($this->cadenasSQL as $xmlorig => $sqlarray) {

                if ($this->resume == 'true')
                  unset($contReferente->sqlExterna2[$xmlorig]);

                foreach($sqlarray as $sql)
                    $contReferente->sqlExterna2[$xmlorig][] = $sql;

                // delete sql statements after
                // unset($this->cadenasSQL[$num]);
            }

        }

          
    // copy eval Statements to parent container
        if ($this->evalStatements != '') {
            foreach ($this->evalStatements as $num => $evalStatement) {

                $contReferente->evalExterna[] = $evalStatement;
            }
        }

      // copy print scripts to parent container
        if ($this->printScripts != '') {
            foreach ($this->printScripts as $num => $pscript) {
                $contReferente->externalPrintScripts[] = $pscript;
            }
        }

        // TODO New Method remove the above one after
        // Upload print containers to parent referent container

        if ($this->printContainerReaders != '')
            foreach ($this->printContainerReaders as $num => $pContainerReader) {
                $contReferente->externalPrintContainerReaders[] = $pContainerReader;
            }

        // refresco referente
        if ($contReferente->xml != '' && $reloadParent) {

            // vuelvo a serializar el contenedor
            Histrix_XmlReader::serializeContainer($contReferente);

            $optionsArray['instance'] = $contReferente->getInstance();

            $postOptions = json_encode($optionsArray);

            $reload = $this->refreshParentScript($postOptions);

            $script .= '<script type="text/javascript">';
            $script .= $reload;
            $script .= '</script>';
        }

        // convertirlo a return o refrescar despues de llamar a esta funcion
        echo $script;
    }

    /**
     * Javascript Refresh parent
     * 
     * @param  string  $postOptions POST params to reload parent
     * @param  boolean $reload      flag to reload parent xml
     * @return string               reload javascript script
     */
    public function refreshParentScript( $postOptions = '', $reload = false)
    {
        if ($postOptions == '') {
            $postOptions = '{instance:\''.$this->parentInstance.'\'}';
        }
        $xmlref = $this->xmlReferente;
        $script = 'grabaABM(\'Form' . $xmlref . '\', \'refresh\', \'' . $xmlref . '\' , \'' . $xmlref . '\', false, null, '.$postOptions.' );';

        if ( isset($this->parentInstance) && $reload == true) {
            $script = 'Histrix.reloadXml($("[instance='.$this->parentInstance.']").closest("div.detalle, div.contenido"));';
        }

        return $script;
    }

    public function saveState(){

        $this->GraboReferente();

    }

    /**
     * Fill parent container data table
     * @return  void 
     */
    public function GraboReferente()
    {
        // Obtengo la tabla temporal de destino

        //Obtengo los Referentes Anexos NOT USED
        //$referentes = $this->referentesAnexos;
        
        $referentes[] = $this->xmlReferente; // agrego el principal

        if ($referentes != '') {
            foreach ($referentes as $numRef => $xmlref) {

                // tengo el contenedor original
                if ($xmlref != '') {

                    $contReferente = new ContDatos("");
                  
                    $contReferente = Histrix_XmlReader::unserializeContainer(null, $this->parentInstance);

                    if (!isset($contReferente->CuerpoMov) && isset($contReferente->parentInstance)){
                      $contReferente = Histrix_XmlReader::unserializeContainer(null, $contReferente->parentInstance);                      
                    }
                    // see if must bubble up to parent container



                }

                // Recorro la tabla temporal e inserto registros en la tabla temporal
                // del contenedor Referente
                //$arraydatos = $this->TablaTemporal->datos();
                $arraydatos = $this->TablaTemporal->Tabla;
                // Recorro el la tabla temporal

                $final = $contReferente->getInsertTemporal();

                $i = 0;
                if ($arraydatos != '') {
                    foreach ($arraydatos as $numTT => $rowTT) {
                        $i++;
                        $arrayin = $contReferente->getInsertTemporal();

                        // Solo traigo del emergente los que tienen este campo en 1
                        // Y si no tengo condiciones a Todos
                        if ($this->CampoCond != '' && 
                        (  $rowTT[$this->CampoCond] == 0
                        || $rowTT[$this->CampoCond] == ''
                        || $rowTT[$this->CampoCond] == false
                        || $rowTT[$this->CampoCond] == 'false' )
                        ) {
                            unset($arrayin);
                            continue;
                        }

                        foreach ($rowTT as $nomcol => $valor) {
                            $exporta = '';
                            $currentField = $this->getCampo($nomcol);
                            $Campdet = $currentField->Detalle;

                            if (isset($Campdet) && $Campdet != '') {
                                $hasReturnData = true;
                                foreach ($Campdet as $ndet => $det) {
                                    //chequeo la existencia del campo
                                    unset($destinationObj);
                                    $destinationObj = $contReferente->getCampo($det);

                                    if (!$destinationObj) {
                                        continue;
                                    }
                                    // get wich field need to be replaced
                                    if ($destinationObj->replace == "true") {
                                        $arrayrep[$det]= "true";
                                    }

                                    if ($this->exportasuma == 'true') {
                                        // Acumulo los campos que se suman
                                        if ($this->seSuma($nomcol)) {
                                            if ($i == 1)
                                                $final[$det] = 0;
                                            $final[$det] += $valor;
                                        } else
                                            $final[$det] = $valor;
                                    }

                                    $arrayin[$det] = $valor;
                                }
                            }
                        }
                        if ($this->exportasuma != 'true' && $hasReturnData == true) {
                            if (isset($arrayin)) {

                                $contReferente->TablaTemporal->insert($arrayin, true, $arrayrep);
                            }
                        }
                        unset($arrayin);
                    }
                }

                if ($this->exportasuma == 'true') {
                    if (isset($final)) {
                        //loger(print_r($final, true), 'debug.log');
                        $contReferente->TablaTemporal->insert($final, true, $arrayrep);
                    }
                }

                $contReferente->calculointerno();

                // save current container into parent medatada
                if ($this->__parent_row != '' && $this->__parent_col != ''){
                
                      $contReferente->TablaTemporal->setMetadata($this->__parent_row, $this->__parent_col, $this);
                }
                // vuelvo a serializar el contenedor
                Histrix_XmlReader::serializeContainer($contReferente);


                $xml = $this->xmlpadre;

                // DO NOT REMOVE COMMENT, BREAKS CLOSE WINDOW ON EXTERNAL
                // if ($this->reloadParent == false) return;

                // refresco referente
                if ($xmlref != '') {
                    $optionsArray['instance'] = '"'.$contReferente->getInstance().'"';
                    $optionsArray['processEvent'] = 'false';
                    $postOptions = Html::javascriptObject($optionsArray, '"');
                    //  $postOptions = htmlspecialchars($postOptions);
                    $refreshEvent = (isset($this->refreshEvent))?$this->refreshEvent:'update';
                    $rel = 'grabaABM(\'Form' . $xmlref . '\', \''.$refreshEvent.'\', \'' . $xmlref . '\' , \'' . $xmlref . '\' , false, null, '.$postOptions.'); ';
                }
                $reload[$rel] = $rel;
                $reload[] = ' cerrarVent(\'DIV' . $xml . '\');';
                $reload[] = ' cerrarVent(\'PRN' . $xml . '\');';

                // DEBERIA EJECUTAR  ESTO????
                // LO SACO EL 30/06/2010 A VER QUE PASA...
                if ($contReferente->xmlOrig != '' && $contReferente->xmlOrig != $xmlref) {

                    //    $rel = 'grabaABM(\'Form' . $contReferente->xmlOrig . '\', \'update\', \'' . $contReferente->xmlOrig . '\' , \'' . $contReferente->xmlOrig . '\' , false); ';
                    //  $reload[$rel] = $rel;
                }

                if ($contReferente->CabeceraMov != '') {
                    foreach ($contReferente->CabeceraMov as $nCabecera => $containerCabecera) {
                        foreach ($containerCabecera->tablas[$containerCabecera->TablaBase]->campos as $numCamCab => $campoCabecera) {
                            if ($campoCabecera->contExterno != '') {
                                if ($campoCabecera->contExterno->xml != '') {
                                    unset($optionsArray);
                                    $optionsArray['instance'] = '"'.$campoCabecera->contExterno->getInstance().'"';
                                    $optionsArray['processEvent'] = 'false';
                                    $postOptions = Html::javascriptObject($optionsArray, '"');

                                    $rel = 'grabaABM(\'Form' . $campoCabecera->contExterno->xml . '\', \'update\', \'' . $campoCabecera->contExterno->xml . '\' , \'' . $campoCabecera->contExterno->xml . '\' , false , null, '.$postOptions.'); ';
                                }
                                $reload[$rel] = $rel;
                            }
                        }
                    }
                }
            }
            // !!!!!SACAR ESTO DE ACA, PONERLO EN EL RETORNO DE LA FUNCION
            // TODO esto arregla los casos de un ing dentro de una fichaing duplica la solapa al procesar y esta MAL
            if ($this->preventDuplication != 'true') {
                $script = '<script type="text/javascript">';
                $script .= implode('; ', $reload);
                $script .= $cerrar;
                $script .= '</script>';
            }

            echo $script;
        }
    }


    private function grabarInternos($campoFicha, $return = false, $unserialize = true, $updateLastId =true)
    {


        if ($campoFicha->contExterno != '') {




            if (   $campoFicha->contExterno->tipoAbm == 'grid' 
                || $campoFicha->contExterno->tipoAbm == 'ing'
                || isset($campoFicha->contExterno->CuerpoMov) // add if container has movements
            ) {
//                  loger('--------- G interno'.$this->xml, 'updatessql.log');
//                Histrix_XmlReader::serializeContainer($this);
                  
//                  loger('--------- GRABO INTERNO'.$campoFicha->contExterno->xml, 'updatessql.log');

                  $contInterno= $campoFicha->contExterno;

                if ($campoFicha->contExterno->xml != '' && $unserialize) {
                    $contInternoTmp = Histrix_XmlReader::unserializeContainer($campoFicha->contExterno);
                    if ($contInternoTmp != false)
                        $contInterno = $contInternoTmp;
                }

                if ($updateLastId) {
                    $contInterno->lastInsertedId = $this->lastInsertedId;
                }
                $contInterno->parentInstance = $this->getInstance();


                // to allow internat <if> to be processed ok
                $contInterno->calculointerno();
                $contInterno->deserializeParent = false;

                unset($contInterno->cadenasSQL);

                $updateResponse= $contInterno->GrabarRegistros('', '', '',  true, $this);

                // rollBackSupport
                if ($updateResponse === -1) {
                    return -1;
                }

                // if this xml fills SQL to the parent or referent container
                // where use the generated SQL statements
                if ($contInterno->llenoReferente == 'true' ) {
                    if ($return) {
                        $tmp  = $contInterno->cadenasSQL;
                        unset($contInterno->cadenasSQL);

                        return  $tmp;
                    }
                      $this->updateParentContainer($contInterno);
                }
// loger('-------- FIN G internos'.$this->xml, 'updatessql.log');
            }


        }


    }

    public function updateParentContainer(&$contInterno){
    
        if ($contInterno->xmlReferente == $this->xml) {

//     loger('update parent'.$this->xml, 'updatessql.log');
        
            if ($contInterno->sqlExterna2 != '') {
                foreach ($contInterno->sqlExterna2 as $xmlsql => $sqlArray) {
                
                  unset($contInterno->cadenasSQL[$xmlsql]);
                    foreach ($sqlArray as $num => $sql1) {
                        $contInterno->cadenasSQL[$xmlsql][] = $sql1;
                    }
                }
                unset($contInterno->sqlExterna2);
                
            }

            if ($contInterno->cadenasSQL) {
                foreach ($contInterno->cadenasSQL as $num => $sqlarray) {
          // remove previous
                  unset($this->sqlExterna2[$num]);
                    foreach($sqlarray as $xmlsql => $sql){
                      $this->sqlExterna2[$num][] = $sql;
                    }
                }
                unset($contInterno->cadenasSQL);
            }
//loger($this->sqlExterna2, 'updatessql.log');


          if (isset($contInterno->lastInsertedId)) {
          if (isset($this->lastInsertedId)) {
                  $this->lastInsertedId = array_merge($this->lastInsertedId, $contInterno->lastInsertedId);
          } else {
                
                  $this->lastInsertedId = $contInterno->lastInsertedId;
                }
           }


            if ($this->evalExterna != '') {
                foreach ($contInterno->evalExterna as $nsql => $ev) {
                    $contInterno->evalStatements[] = $ev;
                }
            }
            if ($contInterno->evalStatements) {
                foreach ($contInterno->evalStatements as $num => $ev) {
                    $this->evalExterna[] = $ev;
                }
            }

        } else {
//     loger('update al ref'.$this->xml, 'updatessql.log');
        
            $contInterno->SqlAlReferente($this);
        }
    }

    /**
     * Toma los Registros de la tabla temporal y crea los Movmientos y la cabecera
     */
    public function GrabarRegistros($numeroMinuta = '', $reempN = '', $reempVal = '', $grabaParams = true, $parentContainer= null)
    {
        $this->lastLogcount += 1;
        // Begin Transaction if xml says so
        if ($this->transaccional == 'true') {

            // if there is not another transaction open yet...
            //if ($_SESSION['transaction'] != 'open') {
              _begin_transaction();

           // }

        }

//        loger('GR'.$this->xml, 'updatessql.log');

        // Grabo la minuta antes que nada
        if ($this->graboasiento == 'true' && $numeroMinuta == '') {
            $numeroMinuta = $this->Minuta->Grabar();
            // error en la Grabación de la minuta
            if ($numeroMinuta === -1) {
    //            _rollback_transaction();
    //            die('<div class="error">ERROR EN LA GRABACION DE LA MINUTA CONTABLE, LA TRANSACCION NO SE REALIZO</div>');
            }
        }

        // Grabacion de los Registros UNICOS que toman los datos de la cabecera y/o
        // de la suma de los campos y NO del valor de la tabla temporal
        if (isset($this->CuerpoMov)) {
            foreach ($this->CuerpoMov as $numMoim => $relationContainer) {
                $postgrab = isset($relationContainer->postGrabacion)?$relationContainer->postGrabacion:'false';

                if ($relationContainer->grabaFinal == 'true' && $postgrab != 'true') {
                    $respuesta = $this->GrabarRegistrosUnicos($relationContainer, $numeroMinuta);

                    // Rollback support
                    if ($respuesta === -1)
                        return -1;

                    Histrix_XmlReader::serializeContainer($this);

                }
            }
        }
        $cadenaSQL = '';
        // creo la tabla temporal con 1 registro
        if ($this->tipoAbm == 'fichaing') {
            Histrix_XmlReader::serializeContainer($this);


            // Grabo los contenedores internos Antes del contenedor principal
            
            if ($this->processFirst != 'true' ){
              foreach ($this->tablas[$this->TablaBase]->campos as $numcampo => $campoFicha) {
                  $updateResponse = $this->grabarInternos($campoFicha, false, true);
                      // Rollbacl support
                      if ($updateResponse === -1)
                          return -1;
              }
            }

            // Duplicado
            //  $arrayin = $this->CargoTablaTemporalDesdeCampos();

            $det = '';

            foreach ($this->tablas[$this->TablaBase]->campos as $numCamp => $campoFing) {

                    // get parameters for fichaing xml
                $this->getParametros($campoFing);

                $valorFing = $campoFing->getValorGen($campoFing->valor, $campoFing->TipoDato);

                if (count($campoFing->opcion) > 0 && $campoFing->TipoDato != "check" && $campoFing->valoropcion == 'true') {
                    $valorFing = $campoFing->opcion[$valorFing];
                    if (is_array($valorFing))
                        $valorFing = current($valorFing);
                }
                $detallecampo = $campoFing->getUrlVariableString($valorFing);
                if ($detallecampo != '') {
                    $det .= '&' . $detallecampo;
                }
            }
            $detallecampo = '';

            $arrayin = $this->CargoTablaTemporalDesdeCampos($parentContainer);
        } 
        //else {
            //  La sentencia SQL generada va a parar al contenedor referente
            if ($this->llenoReferente == 'true')
                $this->GraboReferente();
        //}

        if (isset($this->lastInsertedId)) {
            //            if (isset($contReferente->lastInsertedId)) {
            //                $contReferente->lastInsertedId = array_merge($contReferente->lastInsertedId, $this->lastInsertedId );
            //            }
            //            else {
            //                $contInterno->lastInsertedId = $this->lastInsertedId;
            //            }
        }

        // recorro la tabla temporal
        $ordenTT = 1;
        $arraydatos = $this->TablaTemporal->Tabla;
        // Recorro el la tabla temporal

        if (isset($this->CuerpoMov) && $arraydatos != '')
            foreach ($arraydatos as $numTT => $rowTT) {
            $innerSql = '';

                // Empiezo a recorrer las Filas de la tabla Temporal
      //  $rowTT = array();
                $rowTT["_ORDEN"] = $ordenTT;
//loger($rowTT, 'updatessql.log');
                // Por cada Renglon Recorro el contendedor de la Tablas
                // donde guardar los datos
                // ContDatos de los xml relacionados
                foreach ($this->CuerpoMov as $numMoim => $relationContainer) {

//                  $relationContainer = clone $relationContainer1;

        
        // skip marked containers (this is usefull for processMovemnts)
                    if ($relationContainer->_skip == true)
                        continue;


                    // Si se tiene que grabar al final paso
                    if ($relationContainer->grabaFinal == 'true')
                        continue;

                    $campoCant = $relationContainer->CampoCant;
                    $campoCond = $relationContainer->CampoCond;
                    $campoCondExp = isset($relationContainer->CampoCondExp) ? $relationContainer->CampoCondExp : '';

                    // WARNING do not chang this conditions without PROPER TESTING
                    
                    if ($campoCond != '' && !array_key_exists($campoCond, $rowTT)) {
                      // continue;    THIS BREAKS SOME MALFORMED XMLS --- TEST SOLUTIONS
                      loger('Container: '.$this->xml, 'malformed_xml.log');
                      loger('saves: '.$relationContanier->xml, 'malformed_xml.log');
                    }


                    if (isset($rowTT[$campoCond]) && $campoCond != '' 
                          && ( $rowTT[$campoCond] === 'false' 
                          ||  $rowTT[$campoCond] === false 
                          ||  $rowTT[$campoCond] === 0
                          ||  $rowTT[$campoCond] === '0'
                          ||  trim($rowTT[$campoCond]) === ''   
                           )) {

                        continue;
                    }
                    else {
                  //    loger('si grabo'.$campoCond.'='.$rowTT[$campoCond].'|', 'updatessql.log');
                  //    loger($rowTT,  'updatessql.log');
                    }
                    /*
                  if (trim($rowTT[$campoCond]) === ''){
                    // corner case not working
                      loger('no grabo'.$campoCond, 'no_grabo.log');
                      loger($rowTT, 'no_grabo.log');
                  }
                  */



                  /*
                          if ($rowTT[$campoCond] == 'true') {

                              loger($Nomvim->xml,'condic');

                              loger($campoCond,'condic');
                              loger($rowTT,'condic');
                              }
                    */


                    if ($campoCondExp != '') {
                        $noexec = 'false';
                        $evaluado = null;
                        $objCampoCond = $this->getCampo($campoCond);
                        $valorUltimo = $objCampoCond->ultimo;

                        if (trim($valorUltimo) != '') {

                        } else {
                            if ($objCampoCond != '') {
                                $valorUltimo = $objCampoCond->valor;
                            } else {
                                if (is_object($parentContainer)) {
                                    $valorUltimo = $parentContainer->getCampo($campoCond)->valor;
                                }

                                if (isset($this->CabeceraMov)) {
                                    foreach ($this->CabeceraMov as $nCabecera => $containerCabecera) {
                                        if ($containerCabecera->getCampo($campoCond)) {
                                            $valorUltimo = $containerCabecera->getCampo($campoCond)->valor;
                                        }
                                    }
                                }
                            }
                        }

                        if ($valorUltimo == '') {
                              $valorUltimo = 0;
                        }


                        $expresionaevaluar = '$evaluado = (' . $valorUltimo . ' ' . $campoCondExp . ')?"true":"false";';
                        loger('Evaluo CondExp:');
                        loger($relationContainer->xml . ' ' . $expresionaevaluar, 'updatessql.log');
                        eval($expresionaevaluar);

                        if ($evaluado == 'false') {
                            //$noexec = 'true';
                              loger($relationContainer->xml . ' NO GRABO', 'updatessql.log');
                              continue;
                        } else {
                              loger($relationContainer->xml . ' SI GRABO', 'updatessql.log');
                        }
                    }

                   
                    if ($campoCant == '')
                        $multi = 1;
                    else {
                        $multi = ($rowTT[$campoCant] != '') ? $rowTT[$campoCant] : 1;
                        if (is_array($multi)) {
                            // If this field has a DataContainer inside
                            $filaArray = current($multi[0]);
                            $multi = count($filaArray);
                        }
                    }

                    // Lo hago N veces segun cuantos registros por xml se tengan que grabar
                    // TODO: revisar si este bucle se ha utilizado alguna vez

                    for ($i = 0; $i < $multi; $i++) {

                        $rowTT["__ROWCOUNT__"] = $i + 1;

                        if (isset($filaArray) && is_array($filaArray)) {
                            if ($filaArray[$i + 1] == 'false' || $filaArray[$i + 1] == 0 || $filaArray[$i + 1] == '') {
                                continue;
                            }
                        }

                        // agrego condiciones para los updates
                        // ERROR: ESTOY REVISANDO TODOS LOS CAMPOS DEL TEMPORAL... y si no se usan es inutil
                        foreach ($rowTT as $columna => $campoTT) {

                            $nomMovimObjTemp = $relationContainer->getCampo($columna);
                            if ($nomMovimObjTemp) {
                                if ($nomMovimObjTemp->TipoDato == "date") {
                                    $campoTT = Types::formatDate($campoTT);
                                }
                                $relationContainer->addCondicion($columna, '=', $campoTT);
                                $relationContainer->setNuevoValorCampo($columna, $campoTT);
                            }


                            // save Inner SQL metadata
                            $metadata = $this->TablaTemporal->getMetadata( $numTT , $columna);

                            if ($metadata != '' && $metadata->saveState == 'true') {

//                                loger('metadata', 'updatessql.log');
//                                loger($numTT .'--'. $columna, 'updatessql.log');
                                

                                // if this xml fills SQL to the parent or referent container
                                // where use the generated SQL statements
                                if ($metadata->llenoReferente == 'true' || $metadata->llenoreferente == 'true') {

                                  $metadata->grabarRegistros('','','',true, $this);
                                  $this->updateParentContainer($metadata);
                                  
                                // prevents DOUBLE SAVING
                                  $this->TablaTemporal->destroyMetadata($numTT, $columna);

                                  //$metadata->SqlAlReferente($this);

                                }
                            } 

                        }

                        // Asigno el valor de cada Campo tomandolo de la Tabla temporal
                        // Recorro cada uno de los campos del XML que graba
                        unset ($cadena2SQL);

                        $relationContainer->CargoTablaTemporalDesdeCampos();

                        foreach ($relationContainer->tablas[$relationContainer->TablaBase]->campos as $Ncampo => $campoMovim) {

                            //TODO: TEST USE CASES to see if can be removed
                            // its used in barcode printing
                            if ($campoMovim->param != '') {

                                $ParametroDinamico = $relationContainer->getCampo($campoMovim->param)->ContParametro;

                                // Not_So_Magic_Quotes Value
                                $valorcond = Types::getQuotedValue($campoMovim->getNuevoValor(), $campoMovim->TipoDato, 'xsd:integer');

                                $ParametroDinamico->addCondicion($campoMovim->NombreCampo, '=', $valorcond, 'and', 'reemplazo');
                                unset($valorcond);
                            }

                            // Obtengo el nombre del campo que se utiliza en la asignacion

                            $nomCampoTemp  = $campoMovim->campoTemp;
                            $nomCampoCab   = $campoMovim->campoCab;
                            $nomCampoPadre = $campoMovim->campoPadre;
                            $nomCampoInt   = $campoMovim->campoInterno;
                            $nomCampoCont  = $campoMovim->contenedorInterno;

                            // Retrive Parameters Value
                            if ($relationContainer->getParametros($campoMovim)) {
                                $relationContainer->CargoTablaTemporalDesdeCampos();
                                $relationContainer->calculointerno();
                            }

                            $arrayTT = $relationContainer->TablaTemporal->Tabla;




                            // Obtengo el valor de la Tabla Temporal
                            $valor = $arrayTT[0][$campoMovim->NombreCampo];


                            // force eval on 
                            if ($campoMovim->eval =="true"){

                              try {
                                $codigo = '$valor = ' . $campoMovim->valor . ';';
                                eval($codigo);

                              } catch (Exception $e) {
                                loger($e, 'eval_errors.log');

                              }
                              if (is_object($valor)){
                                $valor = $valor();
                              }
                              
 //                             loger($valor, 'eval code');
                            }




                            if ($nomCampoTemp != '' && array_key_exists($nomCampoTemp, $rowTT)) {

                                if ($campoMovim->TipoDato == "date") {
                                    $valor = Types::formatDate($rowTT[$nomCampoTemp]);
                                } else {
                                    $valor = $rowTT[$nomCampoTemp];

                                    $campoTemp = $this->getCampo($nomCampoTemp);

                                    // ON COMPLETE SET PARAMETERS ONLY IF IS USED
                                    // WARNING THIS MAY BRAKE A LOT OF THINGS
                                    /*
                                     if (isset($campoTemp->ContParametro)) {
                                            $campoMovComplete = $campoTemp->ContParametro->getCampo($campoTemp->ContParametro->retorna);
                                            if (isset($campoMovComplete->onComplete) && $grabaParams) {
                                                $nuevoValor = $campoMovComplete->onComplete;
                                                $cadena2SQL[$nomCampoTemp] = $this->setParametros($campoTemp, $nuevoValor, true);
                                            }
                                    }
                                    */


                                    if ($campoTemp->contExterno != '' && $campoTemp->contExterno->isInner == 'true') {
                                        $campoTemp->contExterno->reloadParent = false;

                                        //$campoTemp->displayInnerTable($campoTemp, $rowTT, $Valcampo, $orden, $x, $y , $parametros);
                                        $campoTemp->refreshInnerDataContainer($this, $rowTT);

                                        $campoTemp->contExterno->esInterno = true;

                                        $UI = 'UI_'.str_replace('-', '', $campoTemp->contExterno->tipo);
                                        $abmDatosDet = new $UI($campoTemp->contExterno);
//                                        loger($campoTemp->contExterno->TablaTemporal, 'inst');

                                        $abmDatosDet->showTablaInt('micro','','','false',true, 'noform' );
                                        Histrix_XmlReader::serializeContainer( $campoTemp->contExterno);
                                        $tempInnerSql = $this->grabarInternos($campoTemp, true, false, false);

                                        // rollback support
                                        if ($tempInnerSql === -1)
                                          return -1;

                                        if (is_array($innerSql))
                                            $innerSql = array_merge($innerSql, $tempInnerSql);
                                        else
                                        $innerSql = $tempInnerSql;

                                        unset($tempInnerSql);

                                        $inlineSaving = true;
                                    }


                                    // if there is a DataContainer inside
                                    if (is_array($rowTT[$nomCampoTemp])) {

                                        if ($campoMovim->innerKey != '') {
                                            $innerContainer = $campoTemp->contExterno;
                                            $innerData = $innerContainer->TablaTemporal->Tabla;
                                            $valor = $innerData[$i][$campoMovim->innerKey];

                                            if ($valor == '') {
                                                $current = $rowTT[$nomCampoTemp][0];
                                                $valor = $current[$campoMovim->innerKey][$i + 1];
                                            }
                                        } else {
                                            $innerArray = current($rowTT[$nomCampoTemp][0]);
                                            $valor = $innerArray[$i + 1];
                                        }
                                    }
                                    if ($campoTemp) {

                                        if (count($campoTemp->opcion) > 0 && $campoTemp->TipoDato != "check" && $campoTemp->valoropcion == 'true') {
                                            $valor = $campoTemp->opcion[$rowTT[$nomCampoTemp]];
                                            if (is_array($valor))
                                                $valor = current($valor);
                                        }
                                    }
                                }

                                // i Should put this allways
                           //     if ($relationContainer->sqlhardcoded == 'true' || $relationContainer->customSelect != '' || $relationContainer->customSql != '') {
                                    $reemplazablesNombres[$nomCampoTemp] = '[__' . $nomCampoTemp . '__]';
                                    $reemplazablesValores[$nomCampoTemp] = addslashes($valor);
                              //  }
                            }

                            // Si los valores salen de la suma de un Contenedor de un campo
                            if ($nomCampoInt != '') {
                                $valor = $this->getInnerValue($nomCampoCont, $nomCampoInt);
                            }

                            // excepcion para la grabacion de los asientos (ver como generalizar despues)

                            if ($this->graboasiento) {
                                if ($campoMovim->esNroMinuta == true)
                                    $valor = $numeroMinuta;
                            }

                            // value is last inserted ID of reference xml

                            if ($campoMovim->lastId != '') {
                                if (isset($this->lastInsertedId[$campoMovim->lastId]) &&
                                      $this->lastInsertedId[$campoMovim->lastId] > 0) {
                                    $valor = $this->lastInsertedId[$campoMovim->lastId];

                                $reemplazablesNombres[$campoMovim->lastId] = '[__' . $campoMovim->lastId . '__]';
                                $reemplazablesValores[$campoMovim->lastId] = $valor;

                                } else {
                                    $valor = '[__lastId__' . $campoMovim->lastId . '__]';
                                }

                            }

                            // Formateo el campo
                            $formatoCampo = $campoMovim->Formato;
                            /* FORMATEO EXPLICITO */
                            if ($formatoCampo != '')
                                $valor = sprintf($formatoCampo, $valor);

                            switch ($campoMovim->TipoDato) {
                                case "custom_numeric" :
                                        $valor = doubleval(str_replace(",","",$valor));
                                break;
                            }

                            // Asigno Valor
                            $relationContainer->setNuevoValorCampo($campoMovim->NombreCampo, $valor);

                            // Seteo propiedades del XML
                            $campoMovim->updateSetters($valor);

                            if ($relationContainer->tipoAbm != 'update') {
                                $relationContainer->setCampo($campoMovim->NombreCampo, $valor);
                            }
                            /*
                            if ($relationContainer->tipoAbm == 'imputacion') {
                                if ($valor != '' && $campoMovim->esClave == 'true')
                                    $relationContainer->addCondicion($campoMovim->NombreCampo, '=', Types :: getQuotedValue($valor, $campoMovim->TipoDato, 'xsd:integer'), 'and', 'reemplazo');
                                else
                                    $relationContainer->setNuevoValorCampo($campoMovim->NombreCampo, $valor);
                            }
                            */

                            /**
                             * Cuando es una consulta anidada busco los datos del contenedor padre
                             */
                            if ($nomCampoPadre != '') {

                                if (!isset($contPadre)) {
                                    // unserialize ONLY 1 TIME
                                    // TEST IF THIS IS EXECUTING JUST ONCE

                                    if ($this->xmlpadre != '') {

                                        $contPadre = new ContDatos("");
                                        $contPadre = Histrix_XmlReader::unserializeContainer( null, $this->parentInstance);

                                    }
                                }

                                $campoPadre = $contPadre->getCampo($nomCampoPadre);
                                $contPadre->getParametros($campoPadre);

                                if ($campoPadre === false) {
                                    echo 'campo inexistente: '.$nomCampoPadre;
                                }

                                if (is_object($campoPadre))
                                    $valorPadre = $campoPadre->getValorGen($campoPadre->valor, $campoPadre->TipoDato);
                                else
                                    $valorPadre = $campoPadre->valor;

                                if (count($campoPadre->opcion) > 0 && $campoPadre->TipoDato != "check" && $campoPadre->valoropcion == 'true') {

                                    $valorPadre = $campoPadre->opcion[$valorPadre];
                                    if (is_array($valorPadre))
                                        $valorPadre = current($valorPadre);
                                }

                                $reemplazablesNombres[$nomCampoPadre] = '[__' . $nomCampoPadre . '__]';
                                $reemplazablesValores[$nomCampoPadre] = addslashes($valorPadre);

                                // Tomo el valor de la cabecera y lo cargo en el registro del Movimiento
                                // Si coninciden las relaciones
                                // Formateo el campo
                                $formatoCampo = $campoMovim->Formato;
                                /* FORMATEO EXPLICITO */
                                if ($formatoCampo != '')
                                    $valorPadre = sprintf($formatoCampo, $valorPadre);

                                $relationContainer->setFieldValue($campoMovim->NombreCampo, $valorPadre, 'both');

                                // Seteo propiedades del XML
                                $valor = $valorPadre;
                                $campoMovim->updateSetters($valor);
                            }

                            /*
                             * Aca tengo que tomar los datos de la Cabecera para agregarlo en los registros
                             * Si es necesario
                             */

                            if (isset($this->CabeceraMov) && $nomCampoCab != '')
                                foreach ($this->CabeceraMov as $nCabecera => $containerCabecera) {
                                    //  $detallecampo='';
                                    // tomo el valor del numerador aca de nuevo
                                    $campoCabecera = $containerCabecera->getCampo($nomCampoCab);

                                    if ($campoCabecera != false && $campoCabecera != '') {
                                        $containerCabecera->getParametros($campoCabecera);

                                        $valorCabecera = $campoCabecera->valor;

                                        if (count($campoCabecera->opcion) > 0 && $campoCabecera->TipoDato != "check" && $campoCabecera->valoropcion == 'true') {

                                            $valorCabecera = $campoCabecera->opcion[$valorCabecera];
                                            if (is_array($valorCabecera))
                                                $valorCabecera = current($valorCabecera);
                                        }

                                // ON COMPLETE SET PARAMETERS ONLY IF IS USED
                                    // WARNING THIS MAY BRAKE A LOT OF THINGS
                                     /*
                                     if (isset($campoCabecera->ContParametro)) {
                                            $campoMovComplete = $campoCabecera->ContParametro->getCampo($campoCabecera->ContParametro->retorna);
                                            if (isset($campoMovComplete->onComplete) && $grabaParams) {
                                                $nuevoValor = $campoMovComplete->onComplete;
                                                $cadena2SQL[$nomCampoCab] = $this->setParametros($campoCabecera, $nuevoValor, true);
                                            }
                                    }
                                    */
                                        $reemplazablesNombres[$campoCabecera->NombreCampo] = '[__' . $campoCabecera->NombreCampo . '__]';
                                        $reemplazablesValores[$campoCabecera->NombreCampo] = addslashes($valorCabecera);

                                        // Tomo el valor de la cabecera y lo cargo en el registro del Movimiento
                                        // Si coninciden las relaciones
                                        // Formateo el campo
                                        $formatoCampo = $campoMovim->Formato;
                                        /* FORMATEO EXPLICITO */
                                        if ($formatoCampo != '')
                                            $valorCabecera = sprintf($formatoCampo, $valorCabecera);

                                        $relationContainer->setFieldValue($campoMovim->NombreCampo, $valorCabecera, 'both');

                                    } else {
                                        // No lo encuentro en la Cabecera tengo que recorrerla internamente
                                        foreach ($containerCabecera->tablas[$containerCabecera->TablaBase]->campos as $numCamCab => $campoCabecera) {
                                            if ($campoCabecera->NombreCampo != $nomCampoCab) {

                                                $valorCabecera = false;
                                                unset($valorCabecera);
                                                // BUSCO EN EL CONTENEDOR INTERNO
                                                if ($campoCabecera->contExterno != '' && $campoCabecera->contExterno->xml != '') {

                                                    // TODO: MODIFY THIS SO WILL ONLY DESERIALIZE ONCE.
                                                    $contExt = new ContDatos("");
                                                    $contExt = Histrix_XmlReader::unserializeContainer($campoCabecera->contExterno);

                                                    if ($contExt->tablas[$campoCabecera->contExterno->TablaBase]->campos != '')
                                                        foreach ($contExt->tablas[$campoCabecera->contExterno->TablaBase]->campos as $numCamExt => $campoExt) {
                                                            if ($campoExt->NombreCampo == $nomCampoCab) {
                                                                $valorCabecera = $campoExt->valor;
                                                                if (isset($campoExt->suma))
                                                                    $valorCabecera = $campoExt->Suma;
                                                                else
                                                                    $valorCabecera = $campoExt->ultimo;
                                                            }
                                                        }
                                                }
                                                if (isset($valorCabecera)) {

                                                    // Formateo el campo
                                                    $formatoCampo = $campoMovim->Formato;
                                                    /// FORMATEO EXPLICITO
                                                    if ($formatoCampo != '')
                                                        $valorCabecera = sprintf($formatoCampo, $valorCabecera);

                                                    $relationContainer->setFieldValue($campoMovim->NombreCampo, $valorCabecera, 'both');

                                                }
                                                unset($valorCabecera);
                                            }
                                        }
                                    }
                                    unset($campoCabecera);
                                }


                            // Detalle de los Nmovim
                            // Se utiliza para Imprimir desde los xml que se procesan.
                            if ($relationContainer->xmlImpresion != '') {
                              loger('arma string 1812');
                                $detallecampoMovim = $campoMovim->getUrlVariableString($valor);
                                if ($detallecampoMovim != '') {
                                    $detNmovim .= '&' . $detallecampoMovim;
                                }
                            }

                        } // Fin del bucle CampoMovim
              //        loger('0', 'sqls');
loger('$detNmovim'.$detNmovim);
                        /*
                         * LO MUEVO DONDE TIENE QUE IR ARRIBA NO UNA VEZ POR FILA
                         */

                        if (isset($this->CabeceraMov) && $relationContainer->tipoAbm != 'imputacion') {
                            $det = '';
                            foreach ($this->CabeceraMov as $nCabecera => $containerCabecera) {
                                $detallecampo = '';
                                // tomo el valor del numerador aca de nuevo
                                foreach ($containerCabecera->tablas[$containerCabecera->TablaBase]->campos as $numCamCab => $campoCabecera) {

                                    $valorCabecera = $campoCabecera->getValorGen($campoCabecera->valor, $campoCabecera->TipoDato);

                                    if (count($campoCabecera->opcion) > 0 && $campoCabecera->TipoDato != "check" && $campoCabecera->valoropcion == 'true') {
                                        $valorCabecera = $campoCabecera->opcion[$valorCabecera];
                                        if (is_array($valorCabecera))
                                            $valorCabecera = current($valorCabecera);
                                    }
                                    $reemplazablesNombres[$campoCabecera->NombreCampo] = '[__' . $campoCabecera->NombreCampo . '__]';
                                    $reemplazablesValores[$campoCabecera->NombreCampo] = addslashes($valorCabecera);

                                    $detallecampo = $campoCabecera->getUrlVariableString($valorCabecera);
                                    if ($detallecampo != '')
                                        $det .= '&' . $detallecampo;

                                    unset($valorCabecera);
                                }
                            }
                        }

                        // Ejecuto los exec del xml
                        // Esto se ejecuta uno por linea
                        
                        if (isset($relationContainer->codigoInsert)) {
                            
                            if ($relationContainer->sqlhardcoded == 'true') {
                                $codigoEvaluado = $relationContainer->calculoExpresion($relationContainer->codigoInsert, $arrayTT[0], null, $reemplazablesNombres, $reemplazablesValores);
                            } else {
                                
                                $relationContainer->CargoTablaTemporalDesdeCampos();
                                $relationContainer->calculointerno();

                                $arrayTT = $relationContainer->TablaTemporal->Tabla;

                                $codigoEvaluado = $relationContainer->calculoExpresion($relationContainer->codigoInsert, $arrayTT[0]);
                            }

                            try {
                              loger($codigoEvaluado, 'eval code');
                              eval($codigoEvaluado);
                            } catch (Exception $e) {
                              loger($e, 'eval_errors.log');
                            }

                            if (isset($code))
                               echo Html::scriptTag($code);

                        }

                        

                        $cadenaSQL = '';
                        /*
                         * Acumulo las sentencias SQL de los inserts
                         */
                        if ($relationContainer->tipoAbm != 'nosql') {

                            switch ($relationContainer->tipoAbm) {
                                case 'update' :
                                    $tipoSQL = 'update';
                                    $cadenaSQL = $relationContainer->getUpdate(true) . ' ';
                                    break;
                                case 'delete' :

                                    $cadenaSQL = $relationContainer->getDelete() . ' ';
                                    break;
                                    /*
                                case 'imputacion' :
                                    // le agrego a este caso especial la cabecera dentro del contenedor
                                    if (isset($this->CabeceraMov)) {
                                        unset($relationContainer->CabeceraMov);
                                        foreach ($this->CabeceraMov as $nCabecera => $containerCabecera) {
                                            $relationContainer->addCabecera($containerCabecera);
                                        }
                                    }

                                    $valorImputar = $rowTT[$this->campoValorImputado];
                                    $this->Imputacion($relationContainer, $valorImputar);
                                    unset($relationContainer);
                                    continue;
                                    break;
                                    */

                                default :
                                    $cadenaSQL = $relationContainer->getInsert() . ' ';
                                    $tipoSQL = 'insert';
                                    break;
                            }
                        } else {

                        // do not generate SQL query
                         //           $cadenaDump = $relationContainer->getInsert();
                         //           loger($cadenaDump, 'dumb');

                        }

                        // Aca Actualizo los valores DE LOS PARAMETROS (Parametros o numeradores)
                        // TEST!!!!!!!!!!!
                     //   unset($cadena2SQL );
                     // si comento esto no funcion comprobante interno de caja, controlar

                        if ($cadena2SQL == '')
                            foreach ($this->tablas[$this->TablaBase]->campos as $numCamMov => $campoMov) {
                                if (isset($campoMov->ContParametro)) {

                                    $campoMovComplete = $campoMov->ContParametro->getCampo($campoMov->ContParametro->retorna);
                                    if (isset($campoMovComplete->onComplete) && $grabaParams) {
                                        $nuevoValor = $campoMovComplete->onComplete;

                                        $cadena2SQL[$numCamMov] = $this->setParametros($campoMov, $nuevoValor, true);
                                    }
                                }
                            }

                        /**
                         * Si viene de un referente entonces las posibles sentencias SQL
                         * Se tienen que ejecutar cuando se confirme el referente y no el referido
                         */

                        $updateResponse = '';
                        if ($cadenaSQL != '' || $cadena2SQL != '') {

                          // replace [__vars__] with value
              //              if ($relationContainer->sqlhardcoded == "true") {
                                $cadenaSQL  = trim(str_replace($reemplazablesNombres, $reemplazablesValores, $cadenaSQL));
                                $cadenaSQL2 = trim(str_replace($reemplazablesNombres, $reemplazablesValores, $cadenaSQL2));
                //            }
                            if ($reempN != '') {
                                $cadenaSQL  = trim(str_replace($reempN, $reempVal, $cadenaSQL));
                                $cadenaSQL2 = trim(str_replace($reempN, $reempVal, $cadenaSQL2));
                            }
                   
                            if (strpos($cadenaSQL, '[__') && strpos($cadenaSQL, '__]')) {

                                $cadenaSQL = trim(str_replace($reemplazablesNombres, $reemplazablesValores, $cadenaSQL));

                                // sanitize SQL before insert  ????? 23/06/2012 coment out!!!
                            //    $cadenaSQL =  preg_replace('/\[__+[a-z0-9._]+__\]/i', '0', $cadenaSQL);

                            }


                            if ($this->llenoReferente == 'true') {

                                $this->cadenasSQL[$relationContainer->xml][] = $cadenaSQL;
                                if ($cadena2SQL != '') {
                                    foreach ($cadena2SQL as $parnum => $cadena2SQLstr) {
                                        $this->cadenasSQL[$relationContainer->xml][] = $cadena2SQLstr;
                                    }
                                }

                                $cadenaSQL = '';
                                $cadena2SQL = '';
                            }


                            // prevent error if not avail lastid
                            if (strpos($cadenaSQL, '[__') && strpos($cadenaSQL, '__]')){
                              loger($cadenaSQL,'error.log');
                              $cadenaSQL = '';
                              
                            }


                            if (trim($cadenaSQL) != '') {
                                $updateResponse = updateSQL($cadenaSQL, $tipoSQL, $this->xml);

                                // added Rollback support
                                if ($updateResponse === -1)
                                    return -1;

                                unset($cadenaSQL);
                            }

                            if ($cadena2SQL != '') {
                                foreach ($cadena2SQL as $parnum => $cadena2SQLstr) {
                                    $updateResponse2 = updateSQL($cadena2SQLstr, $tipoSQL, $this->xml);

                                    // added Rollback support
                                    if ($updateResponse2 === -1)
                                        return -1;
                                }
                            }

  

                            //if ($updateResponse2 === -1 || $updateResponse === -1)
                            //    return -1;

          //////////////////
                            //save lastID, this is important, can be used latar in others xmls
                            /////////////////
                            $this->lastInsertedId[$relationContainer->xml] = $updateResponse;
                        
                            if ($inlineSaving == true) {

                                if (isset($innerSql))
                                foreach ($innerSql as $num => $iSql) {

                                    // Reemplazo en las Sql recibidas los valores de los campos definidos como reemplazables
                                    //  Formato: [__nombrecampo__]
                                    // Last Inserted Id exception
                                    if (isset($this->lastInsertedId)) {
                                        foreach ($this->lastInsertedId as $field => $val) {

                                            $iSql = trim(str_replace('[__lastId__' . $field . '__]', $val, $iSql));
                                        }
                                    }
                                    if ($reempN != '') {
                                        $iSql = trim(str_replace($reempN, $reempVal, $iSql));
                                    }

                                    if (strpos($iSql, '[__') && strpos($iSql, '__]'))
                                        $iSql = trim(str_replace($reemplazablesNombres, $reemplazablesValores, $iSql));

                                    if (trim($iSql) != '')
                                        $updateResponse= updateSQL($iSql, null, $this->xml);

                                    // added Rollback support
                                    if ($updateResponse === -1)
                                        return -1;

                                }

                                // dont unset last insertedId (it will break multiple saves)
                                        //  unset($this->lastInsertedId[$relationContainer->xml]);
                            }
                            $inlineSaving = false;

                        }

                        // seve metadata from inner containers
                        foreach ($rowTT as $columna => $campoTT) {
                            // save Inner SQL metadata
                            $metadata = $this->TablaTemporal->getMetadata( $numTT , $columna);

                            if ($metadata != '' && $metadata->saveState == 'true') {

                                if ($metadata->llenoReferente != 'true') {

                                  if (is_array($metadata->lastInsertedId)){
                                    $metadata->lastInsertedId = array_merge($metadata->lastInsertedId, $this->lastInsertedId);
                                  } else {
                                    $metadata->lastInsertedId = $this->lastInsertedId;
                                  }

                                  $metadata->grabarRegistros('', $reemplazablesNombres,$reemplazablesValores,true, $this);
                                  $this->TablaTemporal->destroyMetadata($numTT, $columna);
                                  

                                }
                            } 
                        }





                        // print por relationContainer
                        // se ejecuta al final de cada proceso del relationContainer para
                        // imprimir dentro de un proceso N veces
                        // a menos que lleno el contenedor referente donde guardo el comando para ejecutarlo al final
                        //
                        // Implementado el metodo de impresion directa

                        if ($relationContainer->xmlImpresion != '') {

                            $printScript = $this->printExternal($relationContainer->xmlImpresion, $detNmovim, $relationContainer->dirImpresion);

                            //loger($relationContainer->xmlImpresion, 'print');
                            //loger($printScript, 'print');
                            // new nethod of direct printing
                            // create Loader for print container

                            $printContainer = $this->printContainer($relationContainer->xmlImpresion, $detNmovim, $relationContainer->dirImpresion);

                            unset($detNmovim);

                            if ($this->llenoReferente == 'true') {
                                $this->printScripts[] = $printScript;

                                $this->printContainerReaders[] = $printContainer;
                            } else {
                                if ($relationContainer->directprint != '') {

                                    $printContainer->printer = $relationContainer->directprint;
                                    $this->externalPrintContainerReaders[] = $printContainer;

                                } else {

                                    echo '<script type="text/javascript">' . $printScript . '</script>';
                                }
                            }

                        }
  

                        if ($this->log == "true") {
                            // new log implementation
                            $this->lastLogcount += 1;
                            if ($this->logReference != '') {
                                $references =  explode(',', $this->logReference);
                                foreach ($references as $reference)
                                    $this->wherelog['__ref__'] .= $this->getCampo($reference)->valor . ' ';
                            }
                            $titlog = ($this->titulo_div != '')?$this->titulo_div:$this->titulo;
                            $loger = new Histrix_Loger($this->xml, $titlog, $this->wherelog, $this->updatelog);
                            $loger->dir = $this->dirxml;
                            $loger->log('insert');
                        }
                    } //fin fro $multi

                    unset($relationContainer);
                } // fin $relationContainer;

                $ordenTT++;
            } //fin por $rowTT;
            // TODO: SACAR ESTO DE ACA Y PONERLO EN EL RETORNO DE LA FUNCION
  
         if ($this->xmlImpresion != '') {

            $dirfile= dirfile($this->xmlImpresion, $this->dirImpresion);
            $this->xmlImpresion = $dirfile['file'];
            $this->dirImpresion = $dirfile['dir'];

            if ($this->dirImpresion != '')
                    $det .= '&dir='.$this->dirImpresion;
            $print = '<script type="text/javascript">';

            $print .= $this->printExternal($this->xmlImpresion, $det, $this->dirImpresion);
            $print .= '</script>';
        }

        /*
         * Aca tengo que Cear la Cabecera de los Movimientos Generados
         */
   
        if (isset($this->CabeceraMov)) {
            foreach ($this->CabeceraMov as $nCabecera => $containerCabecera) {
                foreach ($containerCabecera->tablas[$containerCabecera->TablaBase]->campos as $numCamCab => $campoCabecera) {

                    $valorCabecera = $campoCabecera->valor;
                    if (count($campoCabecera->opcion) > 0 && $campoCabecera->TipoDato != "check" && $campoCabecera->valoropcion == 'true') {
                        $valorCabecera = $campoCabecera->opcion[$valorCabecera];
                        if (is_array($valorCabecera))
                            $valorCabecera = current($valorCabecera);
                    }

                    $reemplazablesNombres[$campoCabecera->NombreCampo] = '[__' . $campoCabecera->NombreCampo . '__]';

                    $valorCabecera = $campoCabecera->getValorGen($valorCabecera, $campoCabecera->TipoDato);

                    $reemplazablesValores[$campoCabecera->NombreCampo] = addslashes($valorCabecera);

                    $campoCabecera->setNuevoValor($valorCabecera);
                    unset($valorCabecera);

                    // Si un campo de la cabecera a su vez tiene ingresos grabo esos registros :)
                    // Esto es mas recursivo que no se que.
                    if ($campoCabecera->contExterno && $campoCabecera->contExterno->xml != '') {
                        if ($campoCabecera->contExterno->tipoAbm == 'ing' || $campoCabecera->contExterno->tipoAbm == 'grid') {
                            $campoCabecera->contExterno = Histrix_XmlReader::unserializeContainer($campoCabecera->contExterno);
                            $campoCabecera->contExterno->parentInstance = $containerCabecera->getInstance();
                            $updateResponse = $campoCabecera->contExterno->GrabarRegistros($numeroMinuta, $reemplazablesNombres, $reemplazablesValores);
                            // rollback Support
                            if ($updateResponse === -1)
                                return -1;

                        }
                    }
                }

                // si se graba la Cabecera en una tabla
                if ($containerCabecera->grabaCab == 'true') {

                    $cadenaSQLCabecera = $containerCabecera->getInsert() . ' ';
                    /**
                     * Si viene de un referente entonces las posibles sentencias SQL
                     * Se tienen que ejecutar cuando se confirme el referente y no el referido
                     */
                    if ($this->llenoReferente == 'true') {
                        $this->cadenasSQL[$containerCabecera->xml][] = $cadenaSQLCabecera;
                        $cadenaSQLCabecera = '';
                    }
                    if (trim($cadenaSQLCabecera) != '')
                        $updateResponse = updateSQL($cadenaSQLCabecera, 'insert', $this->xml);

                    // rollback support
                    if ($updateResponse === -1)
                        return -1;
                }

                // Aca Actualizo los valores de la cabecera (Parametros o numeradores)
                // Recorro los Campos de la cabecera
                foreach ($containerCabecera->tablas[$containerCabecera->TablaBase]->campos as $numCamCab => $campoCabecera) {
                    if (isset($campoCabecera->ContParametro)) {
                        $campoParRet = $campoCabecera->ContParametro->getCampo($campoCabecera->ContParametro->retorna);

                        if (isset($campoParRet->onComplete) && $grabaParams) {
                            $nuevoValor = $campoParRet->onComplete;

                            $containerCabecera->setParametros($campoCabecera, $nuevoValor, false);
                        }
                    }
                }
            }
        }

        if (isset($this->CuerpoMov))
            foreach ($this->CuerpoMov as $numMoim => $relationContainer) {
                if ($relationContainer->grabaFinal == 'true' && $relationContainer->postGrabacion == 'true') {
                    $respuesta = $this->GrabarRegistrosUnicos($relationContainer, $numeroMinuta);
                    // rollback support
                    if ($respuesta === -1)
                        return -1;
                }
            }


        // PROCESS INNER CONTAINERS
        if (isset($this->processFirst) && $this->processFirst == 'true' ){
          foreach ($this->tablas[$this->TablaBase]->campos as $numcampo => $campoFicha) {
              $updateResponse = $this->grabarInternos($campoFicha, false, true);
                  // Rollbacl support
                  if ($updateResponse === -1)
                      return -1;
          }
        }

  

        /* EJECUTO LOS SQL QUE SE GENERARON
         * y borro la Tabla Temporal ?
         * HACERLO COMO TRANSACCION
         */

        //if (isset ($this->cadenasSQL)) {
  //      $this->SqlAlReferente($parentContainer);
          $this->SqlAlReferente();

        //}
        // Grabo las sentencias SQL generadas en el contenedor referido
        // Estas Sentencias vienen de las importaciones o grabaciones hechas por otros contenedores

        if ($this->sqlExterna2 != '') {

           // Merge multiple sql arrays in the correct order
           $newArray='';
           $maxcount = 0;
           foreach ($this->sqlExterna2 as $xmlsql => $sqlArray) {
                  $maxcount = max(count($sqlArray), $maxcount);
           }

           for ($i=0; $i < $maxcount; $i++) {
               foreach ($this->sqlExterna2 as $xmlsql => $sqlArray) {
                  $j++;
                  if ($sqlArray[$i] != '')
                    $newArray[$j][$xmlsql] = $sqlArray[$i];
               }
           }


            // Recorro las SqlExternas recibidas y las voy grabando

            foreach($newArray as $numsql => $sqlArray) {
                foreach ($sqlArray as $xmlsql => $strsql) {

                    // Reemplazo en las Sql recibidas los valores de los campos definidos como reemplazables
                    //  Formato: [__nombrecampo__]
                    // Last Inserted Id exception
                    if (isset($this->lastInsertedId)) {
                        foreach ($this->lastInsertedId as $field => $val) {

                            $strsql = trim(str_replace('[__lastId__' . $field . '__]', $val, $strsql));
                        }
                    }
                    if ($reempN != '') {
                        $strsql = trim(str_replace($reempN, $reempVal, $strsql));
                    }

                    if ($reemplazablesNombres != '') {
                        $strsql = trim(str_replace($reemplazablesNombres, $reemplazablesValores, $strsql));
                    }

                    // Remplazo por el Numero de minuta
                    if (strpos($strsql, "_NRO_MIN") !== false)
                        $strsql = str_replace("_NRO_MIN", $numeroMinuta, $strsql);
                    $strsql = trim($strsql);

                    if ($strsql != '' && $this->llenoReferente != 'true') {
                        $updateResponse = updateSQL($strsql, 'insert', $this->xml);

                        // Rollback support
                        if ($updateResponse === -1)
                            return -1;

                        $this->lastInsertedId[$xmlsql] = $updateResponse;

                    }
                }
            }
        }
        if ($this->evalExterna != '') {
            // eval extern code
            foreach ($this->evalExterna as $num => $evalExt) {
                if ($evalExt != '') {
                    try {
                        loger($evalExt, 'eval_diferida.log');
                        eval($evalExt);
                    } catch (Exception $e) {
                        loger($e, 'eval_errors.log');
                    }

                }
            }
        }

  

        // NEW METHOD
        // ejecuto las sentencias de Impresion del contenedor referido idem sqlExterna pero para impresion
        // PRINT every print containerReader
        if ($this->externalPrintContainerReaders != '') {
            foreach ($this->externalPrintContainerReaders as $num => $Histrix_XmlReader) {

                $printContainer = $Histrix_XmlReader->getContainer();

                // Print Method
                $UI = 'UI_' . str_replace('-', '', $printContainer->tipo);
                $datos = new $UI($printContainer);
                $hide = $datos->show(); // so select can be made;

                $printer = (isset($Histrix_XmlReader->printer))?$Histrix_XmlReader->printer:$xmlReader->parameters['printer'];

                /* NEW OO Printing Method */
                $containerPrinter = new ContainerPrinter($printContainer);
                $containerPrinter->target = 'printer'; // pdf , mail or printer
                $containerPrinter->printerName = $printer;
                $containerPrinter->printContainer();

            }
        }

        if ($this->llenoReferente != 'true'){
          //unset($this->TablaTemporal->Tabla);
          $this->TablaTemporal->emptyTable();
        }
           

  
        if ($print != '')
            echo $print;


          if ($this->llenoReferente != 'true') {

              
              if ($this->xmlpadre != $this->xml) {

                //  if ($this->forceReloadScript == "true") {
//          $reloadScript[] = "xmlLoader('IMP$this->xml', '".http_build_query($this->_getParameters) ."', {title:'$this->tituloAbm'});";
                      // Es un detalle
          
                      if ($this->reloadForm == 'false'){
                          $reloadScript[] = "$('#$this->__inlineid').html('');";
                      } else {
                        if ($this->__inlineid != '') {
                            $reloadScript[] = "$('#$this->__inlineid').load( 'histrixLoader.php?". http_build_query($this->_getParameters) ."', { __inline:'$this->__inline' , __inlineid:'$this->__inlineid'} , function(){
                                Histrix.calculoAlturas('$this->__inlineid');
                                    } );";
                        }
                      }
        //  }

              } else {
     
                  if ($this->subdir != '')
                      $subdir = '&dir=' . $this->subdir;
                  if ($this->xml != '') {

                      $reloadScript[] = "xmlLoader('DIV$this->xml', '$getString&xml=$this->xml$subdir', {title:'$this->tituloAbm', reload:true, instance:'".$this->getInstance()."'});";

                      if ($this->close == 'true') {
                          $reloadScript[] = "cerrarVent('PRN" . $this->xml . "');";
                      }

                  }
              }

//                loger($reloadScript,'reloader');
              echo Html::scriptTag($reloadScript);
          }

        if ($this->close == 'true' || $this->cerrar_proceso == 'true') {
            $reloadall = "cerrarVent('DIV" . $this->xml . "');";
            $reloadall .= "cerrarVent('PRN" . $this->xml . "');";
            $reloadall .= "cerrarVent('PRNDIV" . $this->xml . "');";
            echo '<script type="text/javascript">';
            echo $reloadall;
            echo '</script>';
        }
  

        if ($this->cerrar_proceso == 'true') {
            if ($this->xmlReferente != '') {
                $contPrincipal = Histrix_XmlReader::unserializeContainer(null, $this->parentInstance);

                // Propagate LastInsertedID to caller
                if (isset($this->lastInsertedId)) {
                    if (isset($contPrincipal->lastInsertedId)) {
                        $contPrincipal->lastInsertedId = array_merge($contPrincipal->lastInsertedId, $this->lastInsertedId);
                    } else {
                        $contPrincipal->lastInsertedId = $this->lastInsertedId;
                    }
                }

                $contPrincipal->forceReloadScript = true;
                $respuesta = $contPrincipal->GrabarRegistros();

                // rollback support
                if ($respuesta === -1)
                    return -1;
            }
        }

        if ($this->transaccional == 'true') {
            _end_transaction();
        }
        

        //  loger('------------------- fin GR'.$this->xml, 'updatessql.log');

    }

    /**
     * print external xml
     * 
     * @param  string $det
     * @return string
     */
    public function printExternal($xmlImpresion, $det='' , $dir='')
    {
        $doPrint = true;

        if ($this->directprint != '') {
            $printer = '&printer=' . $this->directprint;
            $directprint = 'true';
        }

        if ($this->copies != '')
            $copies = '&copies=' . $this->copies;

        if ($this->printCondition != '') {
            $printConditionField = $this->getCampo($this->printCondition);
            if ($printConditionField != '') {
                if ($printConditionField->ultimo == 'false') {
                    $doPrint = false;
                }
            }
        }

        $dir = ($dir != '')?$dir:$this->subdir;

        if ($det != '' && $doPrint) {
            $paraimp = 'histrixLoader.php?dir='. $dir.'&xml=' . $xmlImpresion . $det .
                    '&autoprint=true&' . $printer . $copies;

            if (isset($this->email)) {
                $paraimp .= '&automail=true';
            }

            $carga = 'Histrix.printExt(\'' . $paraimp . '\',\'' . $directprint . '\' ); ';

            $print = $carga;

            return $print;
        }
    }

    public function printContainer($xmlImpresion, $det='')
    {
        $doPrint = true;

        if ($this->printCondition != '') {
            $printConditionField = $this->getCampo($this->printCondition);
            if ($printConditionField != '') {
                if ($printConditionField->ultimo == 'false') {
                    $doPrint = false;
                }
            }
        }

        if ($det != '' && $doPrint) {

            parse_str($det, $parameters);

            $dir = ($dir != '')?$dir:$this->subdir;

            $parameters['autoprint'] = 'true';
            $parameters['dir'] = $dir;
            $parameters['copies'] = $this->copies;

            if ($this->directprint != '')
                $parameters['printer'] = $this->directprint;
loger($det, 'printparameters');
            $xmlReader = new Histrix_XmlReader($this->dirXmlPrincipal, $xmlImpresion, null, null, $parameters['dir']);
            $xmlReader->addParameters($parameters);

            return $xmlReader;
        }

        return false;
    }

    public function addJoin($ContJoin, $tipoJoin = 'LEFT')
    {
        $this->Joins[] = $ContJoin;
        $this->tipoJoin[] = $tipoJoin;
    }

    /**
     * Agrega a al array de Cabeceras de Movimientos
     * @param ContDatos $newContCab Contenedor Cabecera
     */
    public function addCabecera($newContCab)
    {
        $this->CabeceraMov[$newContCab->getInstance()] = $newContCab;
    }

    /**
     * Agrega a al array de Contenedores de Movimientos
     */
    public function addCuerpoMov($newCont)
    {
        if ($newCont != '')
            $this->CuerpoMov[] = $newCont;
    }

    /**
     * Agrega a al array de Contenedores de Preconsultas
     */
    public function addPreConsulta($newCont)
    {
        //    if (trim($newCont)!='')
        $this->preConsultas[] = $newCont;
    }

    /**
    * Recorro el contenedor actual y lleno las preconsultas
    */
    public function cargoPreconsultas($fillTempTable=true)
    {
        foreach ($this->preConsultas as $npre => $preconsulta) {
            $preconsulta->delCondiciones();
            $preconsulta->_skip = false;
        }

        if (isset($this->tablas[$this->TablaBase]->campos))
            foreach ($this->tablas[$this->TablaBase]->campos as $miNro => $fieldItem) {

                if (isset($fieldItem->preparam) && $fieldItem->preparam != '') {

                    $valor = $fieldItem->getValor();
                    if ($fieldItem->Condiciones != '') {
                        foreach ($fieldItem->Condiciones as $nomfilter => $filter) {

                            $valor = $filter->valor;
                            $valor = trim($valor, '"');
                            $valor = trim($valor, "'");
                        }
                    }

                    $nombre = $fieldItem->NombreCampo;

                    foreach ($this->preConsultas as $npre => $preconsulta) {

                        if ($fieldItem->preparamCampo != '')
                            $nombre = $fieldItem->preparamCampo;

                        $field = $preconsulta->getCampo($nombre);

                        $quotedValue = $valor;

                        if ($fieldItem->TipoDato != 'date')   // if i quote dates it will break some querys, sorry
                            $quotedValue = Types::getQuotedValue($valor, $fieldItem->TipoDato, 'xsd:integer');


                        if ($field) {
                            if (trim($valor) != '' || ($fieldItem->TipoDato == 'date' && $field->required != "true" )) {
                                $field->valor = $valor;

                                $preconsulta->addCondicion($nombre, $fieldItem->preparamOP, $quotedValue, 'and', 'reemplazo');
                                $preconsulta->_skip = false;

                            } else {
        if ($field->required != 'false')
                                    $preconsulta->_skip = true;
                            }
                        }
                    }
                }
            }
                   //  die();
        if ($fillTempTable ) {
                foreach ($this->preConsultas as $npre => $preConsulta) {
        if ($preConsulta->_skip != true){ 
                        $cargocond = true;

                      $preConsulta->Select();
                  $preConsulta->CargoTablaTemporal();
                  if ($preConsulta->TablaTemporal->Tabla != '') {
                      foreach ($preConsulta->TablaTemporal->Tabla as $orden => $fila) {
                          $this->TablaTemporal->insert($fila);
                          }
      }
        }
                }
        }
        return $cargocond;
    }

    public function addTabla($table)
    {
        $buscoindex = 'true';
        $this->tablas[$table] = new Histrix_Table($table, $buscoindex);
    }

    /**
     * Add a Data Field to the Data Container
     * @param Campo  $campo   Field Object
     * @param string $Exp     sql expresion
     * @param string $Etiq    field label
     * @param string $formato optional format
     * @param string $table   optional Table Name
     * @param string $local   is local field
     */
    public function addCampo($campo, $Exp = '', $Etiq = '', $formato = '', $table = '', $local = '')
    {
        if ($table == '')
            $table = $this->TablaBase;
        $this->tablas[$table]->addCampo($campo, $Exp, $Etiq, $formato, $table, $local);
    }

    /**
     * Cantidad de campos que tiene la tabla para usar en los ABM's'
     * @return integer
     */
    public function cantCampos()
    {
        foreach ($this->tablas as $clave => $valor)
            $cant += $valor->cantCampos();

        return $cant;
    }

    /**
     *  Array de Nombres de los campos que tiene la tabla para usar en los ABM's
     * @return array
     *
     */
    public function nomCampos()
    {
        foreach ($this->tablas as $clave => $valor)
            foreach ($valor->campos as $num => $campo) {
                $nombres[] = $campo->NombreCampo;
            }

        return $nombres;
    }

    public function addFiltro($campo, $operador, $label = '', $valor = '', $opcion = '', $modificador = '', $grupo = '', $deshabilitado = false, $copia = '')
    {
        $this->filtros[] = new filtro($campo, $operador, $label, $valor, $opcion,
                        $modificador, $grupo, $deshabilitado, $copia);
    }

    public function delCondiciones()
    {
        unset($this->paginaActual);
        // vacio las condiciones
        if (isset($this->filtros))
            foreach ($this->filtros as $key => $valor) {
        if (!is_object($this->getCampo($valor->campo))) {
            //header('Error:500 Campo'.$valor->campo);
            header('HTTP/1.1 500 Error: field'.$valor->campo);
            die('<div class="error">Error, Campo '.$valor->campo.'</div>');
                //echo 'Error:500 Campo'.$valor->campo;
            } else
                    $this->getCampo($valor->campo)->delCondiciones();
            }
    }

    public function addCondicion($campo, $operador, $valor, $OpLogico = ' and ', $modificador = '', $controloTipo = 'true', $grupo='', $fixed=false)
    {
        $objCampo = $this->getCampo($campo);

        if (!($objCampo)) {
            return false;
        }
        /* Esta es la parte nueva que le agrega la condicion al campo */

        $premod = '';
        if (isset($this->filtros))
            foreach ($this->filtros as $clave => $val) {
                if ($val->campo == $campo)
                    $premod = $val->modificador;
                $controloTipo = 'true';
            }
        $cntrl = true;
        $cntrl = $objCampo->addCondicion($OpLogico, $operador, $valor, $modificador, $premod, $controloTipo, $grupo, $fixed);

        return $cntrl;
    }

    public function addOpcion($campo, $valor, $Desc, $table = '')
    {
        if ($table == '')
            $table = $this->TablaBase;
        $this->tablas[$table]->addOpcion($campo, $valor, $Desc);
    }

    public function esOrden($campo)
    {
        return $this->getCampo($campo)->orden;
    }


    public function addOrden($campo, $table = null, $ordenaTemporal = false, $orderType='')
    {
        /* si ya existe lo saco */
        $this->orderField[]=$campo;
        
        $objCampo = $this->getCampo($campo);
        if ($objCampo != false) {

            $this->orden[$campo] = $campo;

            // set order Type ASC or DESC
            if ($orderType != ''){
              $objCampo->tipoOrden = $orderType;
            }

            if ($table != null)
                $objCampo->tablaOrden = $table;

            if (isset($objCampo->orden))
                $objCampo->orden = false;
            else
                $objCampo->orden = true;
        }
        //$this->orden = null; // lo limpio

        /*
        foreach ($this->tablas[$this->TablaBase]->campos as $miNro => $field) {
            $nom = $field->NombreCampo;
            if (isset($field->orden))
                $this->orden[$nom] = $nom;
        }
        */
        $this->ordenaTemporal = $ordenaTemporal;
    }


    public function setOrden($campo, $table = '', $orderType='')
    {
        // si ya existe invierto el orden
        $objCampo = $this->getCampo($campo);

        if ($orderType != ''){
          $objCampo->tipoOrden = $orderType;
        }
        if ($objCampo == false) return;
        if (isset($objCampo->orden)) {
            
            if ($objCampo->tipoOrden == 'ASC')
                $objCampo->tipoOrden = 'ASC';
            else
                $objCampo->tipoOrden = 'DESC';
        } else {
            $objCampo->tipoOrden = 'ASC';
        }
        /* Establezco un nuevo Orden */

        $this->orden = null; // lo limpio

        foreach ($this->tablas[$this->TablaBase]->campos as $miNro => $field) {
            $field->orden = null;
        }
        $objCampo->tablaOrden = $table;
        $objCampo->orden = true;
        $this->orden[$campo] = $campo;
    }


    public function addAcumula($campo, $op)
    {
        /* inicializo */
        if ($op == 'true')
            $this->acumulaCampo[$campo] = true;
    }

    public function addSuma($campo, $op)
    {
        /* inicializo */
        if ($op == 'true')
            $this->sumaCampo[$campo] = true;
    }

    /**
     *
     * @param <type> $field
     * @param <type> $value
     * @param String $new   ('current', 'new', 'both')
     */
    public function setFieldValue($fieldName, $value, $new='current')
    {
        $field = $this->getCampo($fieldName);

        if ($field) {

            $tipo = Types::getTypeXSD($field->TipoDato, 'xsd:string');

            // remove commas from numeric values
            if ($tipo == 'xsd:decimal') {
                $value = str_replace(',','',$value);
            }


            if ($new == 'current' || $new == 'both') {
                $field->valor = $value;
            }

            if ($new == 'new' || $new == 'both') {
                $field->nuevovalor = $value;
            }

            if (isset($field->setOriginalValue) && $field->setOriginalValue == 'true'){
                $field->setValorOriginal($value);
            }


        }
    }


    public function setCampo($campo, $valor)
    {
        $field = $this->getCampo($campo);
        if ($field)
            $field->valor = $valor;
    }

    public function restaurarValores()
    {
        foreach ($this->tablas[$this->TablaBase]->campos as $miNro => $field) {
    // IF REMOVED VALAUTO FIELD DO NOT WORK PROPPERLY
    if ($field->valauto !="true" || $field->setOriginalValue=="true")
            $field->restaurarValores();
        }
  
    }

    public function setNuevoValorCampo($campo, $valor)
    {
        $objCampo = $this->getCampo($campo);
        if ($objCampo != false && $objCampo->TipoDato == 'boolean') {
            if ($valor == 1)
                $objCampo->nuevovalor = true;
            else
                $objCampo->nuevovalor = false;
        }
        if ($objCampo)
            $objCampo->nuevovalor = $valor;
    }

    /**
     * Retorna si el campo especificado se Suma en la consulta o no
     */
    public function seSuma($campo)
    {
        if (isset($this->sumaCampo[$this->getCampo($campo)->NombreCampo]))
            return true;
        else
            return false;
    }

    /**
     * Retorna si el campo especificado se Acumula la Suma en la consulta o no
     */
    public function seAcumula($campo)
    {
        if (($this->acumulaCampo[$this->getCampo($campo)->NombreCampo]))
            return true;
        else
            return false;
    }

    /**
     * Ejecuta el Select Asociado al contenedor y
     * llena los valores de los Campos
     */
    public function cargoCampos()
    {
        // Ejecuto el Select
        $this->Select();
        /* if ($this->tipoAbm == 'filemanager') {
          $campos = count($this->tablas[$this->TablaBase]->campos);
          } else { */
        $campos = _num_fields($this->resultSet);
        $i = 0;
        while ($row = _fetch_array($this->resultSet)) {
            $i = 0;
            while ($i < $campos) {
                $i++;
                $campoactual = _field_name($this->resultSet, $i);
                $objCampo = $this->getCampo($campoactual);
                $objCampo->valor = $row[$campoactual];
            }
        }

        return $i;
        /*  } */
    }

    public function CargoCamposDesdeTablaTemporal()
    {
        $table = $this->TablaTemporal->Tabla;

        foreach ($table[0] as $fieldName => $value) {
            $this->getCampo($fieldName)->valor = $value;
        }
    }

    public function CargoTablaTemporalDesdeCampos($newValues=false, $cal=true, $parentContainer=null)
    {


        //unset($this->TablaTemporal->Tabla);
        $this->TablaTemporal->emptyTable();

        foreach ($this->tablas[$this->TablaBase]->campos as $miNro => $field) {
            //$campos_a_Mostrar[] = $field->NombreCampo;
            $valor = (isset($field->valor)) ? $field->valor : '';
            if ($newValues == true)
                $valor = $field->nuevovalor;

            if ($field->TipoDato == "date") {
                if ($valor == '0000-00-00') {
                    $valor = '';
                } else {
                    if (strpos($valor, '__]')) {
                        // dejo el valor
                    } else
                    if (substr($valor, 2, 1) != '/')
                        $valor = Types::formatDate($valor);
                }
            }

            $fila[$field->NombreCampo] = $valor;
            unset($valor);
        }


        $this->TablaTemporal->insert($fila);
        unset($fila);

        $isTableEx = (isset($this->isTable_ex)) ? $this->isTable_ex : '';
        if (isset($this->xmlOrig) && $this->xmlOrig != '' && $this->xmlOrig != $this->xml && $isTableEx != true) {
            //    die('XMLORIG'.$this->xmlOrig.'__'.$this->xml);
            // Only unserialize container if is not an SQL generator container
            $isSqlXml = (isset($this->isSqlXml)?$this->isSqlXml:false);

            if ($parentContainer != null){
                  $this->ContOrig = $parentContainer;
            } else {
              if (!$isSqlXml)
                  $this->ContOrig = Histrix_XmlReader::unserializeContainer(null, $this->parentInstance);

            }



        }


        if ($cal)
            $this->calculointerno();


    }

    /**
    * Fill DataTable with Data From TempTable
    * this overwrite current Data
    */

    private function tempTableToDataTable()
    {

            $tempTable = $this->tempTables[$this->llenoTemporal]->Tabla;
            foreach ($this->tablas[$this->TablaBase]->campos as $Nfield => $field) {
                $Emptydata[$Nfield] = '';
            }

            foreach ($tempTable as $rowIndex => $tempRow) {
                foreach ($Emptydata as $name => $val) {
                    if (!isset($tempTable[$rowIndex][$name]))
                        $tempTable[$rowIndex][$name] = $val;
                }
            }
            $this->TablaTemporal->Tabla = array_values($tempTable);

            unset($this->tempTables);

            $this->calculointerno();
    }

    /**
     * Cargo la Tabla Temporal desde el select SQL del Contendor
     */
    public function CargoTablaTemporal()
    {

        // La borro primero
        //unset($this->TablaTemporal->Tabla);
        $this->TablaTemporal->emptyTable();

        $num = 0;


        // Fill DataTable with Data From TempTable
        // this overwrite current Data
        if (isset($this->llenoTemporal) && isset($this->tempTables[$this->llenoTemporal])) {

            $this->tempTableToDataTable();

            return true;
        } else {
         // die('sin temp table'.$this->xml);
        }
        /*
         * Si existe una preconsulta cargo la tabla temporal con el contenido de la misma
         */
        if (isset($this->preConsultas) && $this->preConsultas != '') {
            $this->cargoPreconsultas(true);
        }


        $rs = $this->resultSet;

        // remove resultSet if data came from other sources
        if ($this->selectCode != '') unset($rs);
        if ($this->dataSource != '') unset($rs);


        $num = 0;
        if (isset($rs)) {
            while ($row = _fetch_array($rs)) {

                $fila = '';

                foreach ($row as $etiqcampo => $valor) {
                    // SOLO BUSCO LOS CAMPOS EN EL PRIMERO REGISTRO
                    if ($num == 0) {

                        $fieldName = $etiqcampo;
                        /*
                        if ($this->tablas[$this->TablaBase] != '' && isset($this->tablas[$this->TablaBase]->etiquetas_reverse[$etiqcampo])) {
                            $fieldName = $this->tablas[$this->TablaBase]->etiquetas_reverse[$etiqcampo];
                        }
                        */

                        if ($fieldName != '') {
                            if (!isset($localObj[$fieldName])) {
                                $localObj[$fieldName] = $this->getCampo($fieldName);
                            }
                            $objetoCampo = $localObj[$fieldName];
                        } else {
                            if (!isset($localObj[$etiqcampo])) {
                                $localObj[$etiqcampo] = $this->getCampo($etiqcampo);
                            }
                            $objetoCampo = $localObj[$etiqcampo];
                        }

                        $nombreCampo = $objetoCampo->NombreCampo;
                        $sqlFieldName[$etiqcampo] = $nombreCampo;
                        if (isset($objetoCampo->valop) && $objetoCampo->valop == 'true') {
                            $fieldOption[$etiqcampo] = $objetoCampo->opcion;
                        }
                    }

                    $nombreCampo = $sqlFieldName[$etiqcampo];

                    $fila[$nombreCampo] = $valor;

                    if (isset($fieldOption[$etiqcampo])) {
                        $valoropcion = $fieldOption[$etiqcampo][$valor];
                        if (is_array($valoropcion))
                            $valoropcion = current($valoropcion);
                        $fila[$nombreCampo] = $valoropcion;
                    }


                    if (isset($localObj[$sqlFieldName[$etiqcampo]]->aletras) &&
                            $localObj[$sqlFieldName[$etiqcampo]]->aletras == true) {
                        if (is_numeric($valor))
                            $fila[$nombreCampo] = NumeroaLetras($valor);
                    }

                    $maxSizes[$nombreCampo] = max(strlen($valor), $maxSizes[$nombreCampo]);
                }

                $this->TablaTemporal->insert($fila);
                $num++;
            }

        } else {
            
            /////////////////////////////////////
            // fill table with alternate method
            // custom select Code from php
            /////////////////////////////////////

            if ($this->selectCode != '') {
                $data = $this->getDataFromCode();
            }

            /////////////////////////
            // external data Source
            /////////////////////////

            if (isset($this->dataSource)) {
                $dataSource = $this->dataSource;
                if ( method_exists( $dataSource, 'Select')){
                  $data = $dataSource::Select($this);
                } 
            }

            // process Data
            if(is_array($data)) foreach ($data as $key => $row) {
              $this->TablaTemporal->insert($row);
            }
            
        }

        if (isset($maxSizes)) {
          foreach ($maxSizes as $nom => $val) {
              $this->getCampo($nom)->Tammax = $val;
          }
          unset($maxSizes);
        }

        unset($fieldOption);
        unset($sqlFieldName);

        // calculate table
        $this->calculointerno();

        // Reset cursor position
        if (is_object($rs)){
          _fetch_row($rs, 0);
        }
            


    }

    /**
    * Get data from custom code (EVIL eval function used!)
    */
    private function getDataFromCode(){
      $code = (string) $this->selectCode;

      try {
        $data = eval($code);
      } catch (Exception $e) {
        loger($e, 'eval_errors.log');
      }
      //var_dump($data);
      return $data;
    }




    public function Select($opcion = null)
    {

        // prevent Select if alternate Datasource is detected
        if ($this->dataSource != '') return;


        /* Si hago un select general */
        if (count($this->tablas[$this->TablaBase]->campos) < 1) {
            $fieldArray = $this->TablaBase->getFieldMetadata();
            foreach ($fieldArray as $Nro => $fieldItem)
            //                if (isset ($fieldItem["COLUMN_NAME"]));
            //            $this->TablaBase->addCampo($fieldItem["COLUMN_NAME"]);
            //          drop progress support
                if (isset($fieldItem["Field"])
                    );
            $this->TablaBase->addCampo($fieldItem["Field"]);
        }


        $str = $this->getSelect();

        // Primero busco en este contenedor
        if (strpos($str, '[__') && strpos($str, '__]')) {
            $matrizReemplazable = $this->obtengoMatrizReemplazable();
            foreach ($matrizReemplazable as $nombre => $valor) {
                $matrizNombres[] = $nombre;
                $matrizValores[] = $valor;
            }
            $str = str_replace($matrizNombres, $matrizValores, $str);
        }

        // Si sigue busco en el contenedor padre
        if (strpos($str, '[__') && strpos($str, '__]')) {

            if ($this->xmlOrig != '' && $this->xml != $this->xmlOrig) {
                if ($this->ContOrig == '') {
                    $this->ContOrig = Histrix_XmlReader::unserializeContainer(null, $this->parentInstance);
                }
                if ($this->ContOrig != '') {
                    $matrizReemplazable = $this->ContOrig->obtengoMatrizReemplazable();
                    foreach ($matrizReemplazable as $nombre => $valor) {
                        $matrizNombres[] = $nombre;
                        $matrizValores[] = $valor;
                    }
                }
                $str = str_replace($matrizNombres, $matrizValores, $str);

                // Y en la cabecera del padre (que quilombo tengo que rehacer todo esto)
                if (strpos($str, '[__') && strpos($str, '__]')) {
                    if (isset($this->ContOrig->CabeceraMov))
                        foreach ($this->ContOrig->CabeceraMov as $nCabecera => $containerCabecera) {

                            $matrizReemplazable = $containerCabecera->obtengoMatrizReemplazable();

                            foreach ($matrizReemplazable as $nombre => $valor) {
                                $matrizNombres[] = $nombre;
                                $matrizValores[] = $valor;
                            }
                            $str = str_replace($matrizNombres, $matrizValores, $str);
                        }
                }
            }
        }

        //loger($str, 'ut2');

        //$str = utf8_decode($str);
        //loger($str, 'ut2');


        $subdir = (isset($this->subdir)) ? $this->subdir : null;
        $xmlvar = null;
        if (isset($this->xml)) {
            $xml = $this->xml;
            $xmlOrig = (isset($this->xmlOrig)) ? $this->xmlOrig : null;
            $xmlvar = $xmlOrig . '_' . $xml;
        }

        // save last Select on session var
        if (isset($_SESSION['EDITOR']) && $_SESSION['EDITOR'] == 'editor'){
          $this->_lastSelect = $str;
        }
          

        $this->resultSet = consulta($str, '', $opcion, $xmlvar, $subdir);


        if ($this->tipoAbm != 'imputacion') {
            if ((isset($this->paginar) && $this->paginar != '') || (isset($this->detallado) && $this->detallado == 'false')) {
                $rs = consulta('SELECT FOUND_ROWS()', null, 'nolog');
                $row = _fetch_row($rs);
                $this->TotalRegistros = $row[0];
            }
        }

        return $this->resultSet;
    }

    public function camposaMostrar()
    {
        $ksort=false;
        foreach ($this->tablas[$this->TablaBase]->campos as $miNro => $field) {
            if (isset($field->order)) {
                $campos_a_Mostrar[$field->order] = $field->NombreCampo;
                $ksort=true;
            } else
            $campos_a_Mostrar[] = $field->NombreCampo;
        }
        if ($ksort)
            ksort($campos_a_Mostrar, 0);

        return $campos_a_Mostrar;

    }

    public function getJoins()
    {
        $sqlJoin = '';
        /* CAMPOS DE LOS JOINS */
        if (isset($this->Joins))
            foreach ($this->Joins as $nro => $contenedorJoin) {
                $strCondicJoin = '';
                if ($contenedorJoin->grupo != '') {
                    if ($this->grupoJoin[$contenedorJoin->grupo] != 'true')
                        continue;
                }

                if (isset($contenedorJoin->tablas[$contenedorJoin->TablaBase]->campos))
                    foreach ($contenedorJoin->tablas[$contenedorJoin->TablaBase]->campos as $miNro => $fieldItem) {
                        $strcampo = $fieldItem->getSelectSQL($contenedorJoin->TablaBase);

                        if ($strcampo != '') {

                            if ($contenedorJoin->alias != '')
                                $condicionesCampo = $fieldItem->getCondiciones($contenedorJoin->alias);
                            else
                                $condicionesCampo = $fieldItem->getCondiciones($contenedorJoin->TablaBase);

                            if ($strCondicJoin && $condicionesCampo)
                                $strCondicJoin .= ' and ';
                            $strCondicJoin .= $condicionesCampo;
                        }
                    }

                $sqlJoin .= ' ' . $this->tipoJoin[$nro] . ' JOIN ' . $contenedorJoin->TablaBase;
                if ($contenedorJoin->alias != '')
                    $sqlJoin .= ' as ' . $contenedorJoin->alias;
                if ($strCondicJoin != '') {
                    $sqlJoin .= ' on ';
                    $sqlJoin .= $strCondicJoin;
                }
            }

        return $sqlJoin;
    }

    public function getSelect()
    {
        $first = true;
        $strCondic = '';
        $strHaving = '';
        $campos = '';
        if (isset($this->tablas[$this->TablaBase]->campos))
            foreach ($this->tablas[$this->TablaBase]->campos as $miNro => $fieldItem) {

        if (isset($fieldItem->valor))
                    $fieldItem->updateSetters($fieldItem->valor);
                $strcampo = $fieldItem->getSelectSQL($this->TablaBase);
                if ($strcampo != '') {
                    // modifico la cadena de las expresione para que pueda tomar valores ingresados
                    // Se utiliza en Subselect.
                    //$strnew = $this->calculoExpresion($strcampo);
                    //if (!($fieldItem->esOculto())) { TIENE QUE ESTAR COMENTADO
                    // TODO REWRITE THIS TO AVOIND DOUBLE NEGATIVE
                    $select = true;
                    if (isset($fieldItem->noselect) && $fieldItem->noselect == "true") {
                        $select = false;
                    }

                    if ($select) {
                        if ($first == false)
                            $campos .= ' , ';
                        else
                            $first = false;
                        $campos .= $strcampo;
                    }
                    $condicionesCampo = '';
                    $tabla = $this->TablaBase;

                    if (isset($fieldItem->TablaPadre) && $fieldItem->TablaPadre != $this->TablaBase && $fieldItem->TablaPadre != '')
                        $tabla = $fieldItem->TablaPadre;
                    if (isset($fieldItem->local) && $fieldItem->local != 'true') {

                    } else
                        $condicionesCampo = $fieldItem->getCondiciones($tabla);

                    // Si es una expresion la condicion va en el having (a menos que explicitamente no quiera (con el having == false)*/
                    if (isset($fieldItem->Expresion) && $fieldItem->having != 'false') {
                        if ($strHaving && $condicionesCampo)
                            $strHaving .= ' and ';
                        $strHaving .= $condicionesCampo;
                    } else {
                        if ($strCondic && $condicionesCampo)
                            $strCondic .= ' and ';
                        if ($condicionesCampo != '')
                            $strCondic .= '(' . $condicionesCampo . ')';
                    }
                }
            }

        if ($campos == '')
            $campos = '*';

        $nomTabla = $this->tablas[$this->TablaBase]->getNombre();
        $database = (isset($this->database)) ? $this->database . '.' : '';


        $paginar = false;
        if (isset($this->paginar) && $this->paginar != '')
            $paginar = true;

        $clausula = '';
        if ($paginar)
            $clausula = ' SQL_CALC_FOUND_ROWS ';

        $sql = "SELECT " . $clausula . $campos;
        if ($nomTabla != '')
            $sql .= " from " . $database . $nomTabla;

        // ACA VAN LOS JOINS

        $sqlJoin = $this->getJoins();

        //if (isset ($sqlJoin))
        $sql .= $sqlJoin;

        if ($strCondic) {
            $sql .= ' where ';
            $sql .= $strCondic;
        }
        if (isset($this->group) && $this->group != '') {
            $sql .= ' group by ';

            $pri = true;
            foreach ($this->group as $miNro => $fieldItem) {

                if (!($pri))
                    $sql .= ' , ';

                //  if ((strpos($fieldItem, '-'))) $fieldItem = "'".$fieldItem.'"';

                $sql .= $fieldItem;
                $pri = false;
            }
        }
        /* ACA METO EL HAVING */
        if ($strHaving) {
            $sql .= ' having ';
            $sql .= $strHaving;
        }

        if (isset($this->orden)) {
            $sql .= ' order by ';
            $pri = true;
            foreach ($this->orden as $miNro => $fieldItem) {
                $objCampo = $this->getCampo($miNro);
                if (!($pri))
                    $sql .= ' , ';
                $campoOrden = $this->orden[$miNro];

                // DROP PROGRESS SUPPORT
                //if ((strpos($campoOrden, '-')))
                //  $campoOrden = "'" . $campoOrden . '"';

                if (isset($objCampo->tablaOrden) && $objCampo->tablaOrden != '')
                    $campoOrden = $objCampo->tablaOrden . '.' . $campoOrden;

                $ord = $campoOrden;

                $tipoOrden = '';
                if (isset($objCampo->tipoOrden)) {
                    $tipoOrden = $objCampo->tipoOrden;
                    $ord .= ' ' . $tipoOrden;
                }



                if (isset($objCampo->Expresion))
                    $ord = $objCampo->Expresion . ' ' . $tipoOrden;

                if (isset($objCampo->useAlias))
                    $ord = $objCampo->Etiqueta . ' ' . $tipoOrden;


                $sql .= $ord;
                $pri = false;
            }
        }

        // Paginacion
        if ($paginar) {

            $paginateStart = $this->paginaActual * ($this->paginar);
            $paginateEnd = $this->paginar;
            $sql .= ' LIMIT ' . $paginateStart . ',' . $paginateEnd;
        }
        if (isset($this->limit) && ($this->limit > 0)) {
            $sql .= ' LIMIT ' . $this->limit;
        }
        if (isset($this->union['distinct']) && $this->union['distinct'] != '') {
            $sql = '(' . $sql . ') UNION DISTINCT (' . $this->union['distinct'] . ')';
        }


        if (isset($this->unionContainers) && $this->unionContainers != '') {
            $sql = '('.$sql.')';
            foreach ($this->unionContainers as $num => $Union) {
                $sqlUnion = $Union->getSelect();
                $sql .=  ' UNION DISTINCT (' . $sqlUnion . ')';
            }
	    if ($ord != '') {
		$sql .= ' order by '.$ord;
	    }
        }

        return $sql;
    }

    public function obtengoMatrizReemplazable()
    {
        foreach ($this->tablas[$this->TablaBase]->campos as $miNro => $fieldItem) {
            $value = (isset($fieldItem->valor)) ? $fieldItem->valor : '';
            $matrizReemplazable['[__' . $fieldItem->NombreCampo . '__]'] = $value;
        }
        if (isset($this->lastInsertedId) && $this->lastInsertedId != '') {
            foreach ($this->lastInsertedId as $field => $val) {
                $matrizReemplazable['[__lastId__' . $field . '__]'] = $val;
            }
        }

        return $matrizReemplazable;
    }

    public function obtengoMatricesValores($fila = '', $nfila = '', $precalc = false)
    {
        // Ya precalcule la cabecera y los internos
        if (isset($this->Mnoms) && $this->Mnoms != ''
      //   || $precalc == true
          ) {
            $matrizValores = $this->Mvals;
            $matrizNombres = $this->Mnoms;

            foreach ($matrizValores as $key => $value) {
              $valores[ $matrizNombres[$key] ] = $value;
            }

        } else {

            // Valores de los campos de la cabecera
            if ($this->tipoAbm != 'ayuda' && isset($this->CabeceraMov)) {
                    foreach ($this->CabeceraMov as $nCabecera => $containerCabecera) {
                        // desserializo para estar seguro que esta actualizado

                        foreach ($containerCabecera->tablas[$containerCabecera->TablaBase]->campos as $numCamCab => $campoCabecera) {
//                            $matrizNombres[] = $campoCabecera->NombreCampo;
//                            $matrizValores[] = $campoCabecera->valor;

                            $valores[$campoCabecera->NombreCampo] = $campoCabecera->valor;

                            if ($campoCabecera->contExterno != '' && $campoCabecera->contExterno->xml != '') {

                                $contInterno = $campoCabecera->contExterno;
                                $contInterno = Histrix_XmlReader::unserializeContainer($campoCabecera->contExterno);


                                if ($contInterno) {
                                    foreach ($contInterno->tablas[$contInterno->TablaBase]->campos as $numCamIn => $innerField) {


                                        if ($innerField->Suma != '') {
                                            $value = $innerField->Suma;
                                        } else {
                                            $value = $innerField->valor;
                                        }

//                                        $matrizValores[] = $value;
                                        //$matrizNombres[] = $innerField->NombreCampo;

                                        $valores[ $innerField->NombreCampo ] = $value;

                                    }
                                }
                                unset($contInterno);
                            }
                        }
                    }
                
            }

            // Valores de los campos de los contenedores Padre
            if ($this->tipoAbm != 'ayuda' 
              && !isset($this->codigoInsert)  
              && $this->unserializeParent != 'false'
              && $precalc === false
              ) {

                if ((isset($this->xmlOrig) && $this->xmlOrig != '' && $this->xmlOrig != $this->xml) ||
              (isset($this->xmlpadre) && $this->xmlpadre != '' && $this->xmlpadre != $this->xml)
                ) {

                if ((isset($this->xmlpadre) && $this->xmlpadre != '' && $this->xmlpadre != $this->xml) && $this->ContOrig == '') {

                    $this->ContOrig = Histrix_XmlReader::unserializeContainer(null ,$this->parentInstance);
                 }

                    if ($this->ContOrig != '') {
                        $contOrig = $this->ContOrig;

                        foreach ($contOrig->tablas[$contOrig->TablaBase]->campos as $numCamOrig => $campoOriginal) {
                             if ($campoOriginal->public == 'true') {


                              if (isset($campoOriginal->suma)) {
                                  $value = $campoOriginal->Suma;
                              } else {
                                  $value = $campoOriginal->valor;
                              }

//                              $matrizValores[] = $value;
//                              $matrizNombres[] = $campoOriginal->NombreCampo;

                              $valores[$campoOriginal->NombreCampo] = $value;

                            }
                        }
                        unset($contOrig);
                    }
                }
            }
        }

        // Valores de los campos de la tabla principal (reemplazan a los de las cabeceras si tienen el mismo nombre)????
        if ($fila != '')
            foreach ($fila as $columna => $dato) {

                // pongo a 0 los valores numericos vacios

                if ($dato == '') {
                    //$tipo = $this->getCampo($columna)->TipoDato;
                    if (isset($this->tablas[$this->TablaBase]->campos[$columna]))
                        $dataType = $this->tablas[$this->TablaBase]->campos[$columna]->TipoDato;
                    $tipo = Types::getTypeXSD($dataType, 'xsd:integer');

                    if ($tipo == 'xsd:integer' || $tipo == 'xsd:decimal')
                        $dato = 0;
                }

                $valores[$columna] = $dato;

            }

        if ($nfila != '' && !isset($valores['_ORDEN'])) {
//            $matrizNombres[] = '_ORDEN';
//            $matrizValores[] = $nfila;
            $valores['_ORDEN'] = $nfila;
        }


        foreach ($valores as $name => $value) {
            $matrizNombres2[] = $name;
            $matrizValores2[] = addslashes($value);
        }

        $matrices['nombres'] = $matrizNombres2;
        $matrices['valores'] = $matrizValores2;

        // free memory
        unset($valores);
        unset($matrizNombres2);
        unset($matrizValores2);

        return $matrices;
    }






    public function jQueryCalculationString($field)
    {
        $calculateStrings = $field->calculateStrings;


        if ($calculateStrings == '')
            return;

        $namePairs = $this->getUID();
        $comma = '';
        if ($namePairs != '')
            foreach ($namePairs as $name => $uid) {
                $variables.= $comma . $name . ": $('#$uid')";
                $comma = ' , ';
            }
        if ($calculateStrings != '')
            foreach ($calculateStrings as $index => $jsString) {


                $jQueryString .= "$('#" . $field->uid . "').calc( ";
                // fomula
                $jQueryString .= '"' . $jsString . '", ';
                // Variables
                $jQueryString .= ' { ' . $variables . ' }, ';
                // formating Callback
                $tipo = Types::getTypeXSD($field->TipoDato, 'xsd:integer');
                switch ($tipo) {
                    case 'xsd:integer':
                        $jQueryString .= ' function (s){ return parseInt(s);}, ';
                        break;
                    case 'xsd:decimal':
                        $jQueryString .= ' function (s){ return s.toFixed(2);}, ';
                        break;
                    default:
                        $jQueryString .= ' function (s){ return s;}, ';
                        break;
                }
                // Finish Callback...
                $jQueryString .= ' null '; /// for now
                $jQueryString .= '); '; /// for now
            }

        return $jQueryString;
    }

    public function getUID()
    {
        foreach ($this->tablas[$this->TablaBase]->campos as $name => $field) {
            $fields[$name] = $field->uid;
        }

        return $fields;
    }

    // Calculo Expresion
    // Rehacer para que no se recorra la cabecera a cada llamado de la funcion
    public function calculoExpresion($formula = '', $fila = '', $nfila = '', $arrayKeys = '', $arrayValues = '')
    {
        if ($arrayKeys == '' && $arrayValues == '') {
            if (isset($this->CabeceraMov))
                foreach ($this->CabeceraMov as $nCabecera => $containerCabecera) {
                    // desserializo para estar seguro que esta actualizado
                    $this->CabeceraMov[$nCabecera] = Histrix_XmlReader::unserializeContainer($containerCabecera);
                }
            if ($this->xmlOrig != '' && $this->xmlOrig != $this->xml) {

                $this->ContOrig = Histrix_XmlReader::unserializeContainer($this->parentInstance);
            }
            $matrices = $this->obtengoMatricesValores($fila, $nfila);
            $arrayKeys = $matrices['nombres'];
            $arrayValues = $matrices['valores'];
        }

        foreach ($arrayKeys as $order => $key) {
          $newArray[strlen($key)][$key] = $arrayValues[$order];
        }

        // order array by srllen with out using closures (time expensive)
        krsort($newArray);

        foreach ($newArray as $inner ) {
            foreach ($inner as $key => $value) {
              $rOrderKeys[]   = $key;
              $rOrderValues[] = $value;
              
            }
            
        }


/*
        // new implementation will fix inner replaces BUT MAY brake somethig
        $arrayCombined = array_combine($arrayKeys, $arrayValues);

        // sort Array with long variable names first to avoid replace superposition
        uksort($arrayCombined, function ($a,$b){ return mb_strlen($b) - mb_strlen($a);  } );

        $arrayKeys   = array_keys($arrayCombined);
        $arrayValues = array_values($arrayCombined);

*/
        // end of fix for inner replaces




        $resultado = str_replace($rOrderKeys, $rOrderValues, $formula);

        return $resultado;
    }

    public function ejecutoExpresionSql($Field, $fila)
    {
        $Expresion = $Field->Expresion;

        foreach ($fila as $NombreColumna => $valor) {
            $nombres[] = '[__' . $NombreColumna . '__]';
            $valores[] = $valor;
        }

        // this will hold current repeat value in loops
        if (isset($Field->__repeat__) && $Field->__repeat__ >= 0){
            $nombres[] = '__repeat__';
            $valores[] = $Field->__repeat__;

            $Expresion = str_replace('__repeat__', $Field->__repeat__, $Expresion);

            $Expresion = str_replace($nombres, $valores, $Expresion);
        }

        $sql = str_replace($nombres, $valores, $Expresion);


        if (isset($this->CabeceraMov)) {
            foreach ($this->CabeceraMov as $nCabecera => $containerCabecera) {
                foreach ($containerCabecera->tablas[$containerCabecera->TablaBase]->campos as $numCamCab => $campoCabecera) {
                    $valorCabecera = $campoCabecera->getValorGen($campoCabecera->valor, $campoCabecera->TipoDato);

                    $reemplazablesNombres[$campoCabecera->NombreCampo] = '[__' . $campoCabecera->NombreCampo . '__]';
                    $reemplazablesValores[$campoCabecera->NombreCampo] = addslashes($valorCabecera);
                }
            }
            $sql = str_replace($reemplazablesNombres, $reemplazablesValores, $sql);
        }
        if ($sql != '')
            $rs = consulta($sql);

        if (is_object($rs))
            $row = _fetch_row($rs);
        $val = $row[0];

        return $val;
    }

    /**
     * calculate field values
     * @param  array  $fila    row
     * @param  integer  $nfila   row number
     * @param  integer  $cont    row number again?
     * @param  boolean $precalc true if parent containers have been previously deserialized
     * @return boolean           success
     */
    public function calculoInternoFila($fila, $nfila, $cont, $precalc = false)
    {
        $valid = 1;

        foreach ($fila as $columna => $dato) {
            $fieldObject = &$this->getCampoRef($columna);

            if ($fieldObject == null)
                continue;


            $fila2 = $this->TablaTemporal->Tabla[$nfila];
            $val = $dato;
            $retroalimenta = (isset($fieldObject->retroalimenta)) ? $fieldObject->retroalimenta : '';

            // Sum values
            if (isset($fieldObject->suma) && $fieldObject->suma == 'true' && $retroalimenta != 'true') {
                if ($cont == 0)
                    $fieldObject->Suma = 0;
                $val = $this->TablaTemporal->Tabla[$nfila][$columna];
                $fieldObject->Suma += $val;
                $fieldObject->updateSetters($fieldObject->Suma);
            }

            // set previuos Value
            if (isset($objetoCampo->valop) && $fieldObject->previousValue != '') {
                if ($nfila - 1 >= 0) {
                    $val = $this->TablaTemporal->Tabla[$nfila - 1][$fieldObject->previousValue];
                    $this->TablaTemporal->Tabla[$nfila][$columna] = $val;
                }
            }

            // IF
            if (isset($fieldObject->ifs)) {

                // Solo va a entrar una vez por Campo y no 3
                $matrices = $this->obtengoMatricesValores($fila2, $nfila, $precalc);
                //loger('obtengo',$fieldObject->NombreCampo.'.if');

                if (isset($fieldObject->__repeat__)){
                  $matrices['nombres'][]='__repeat__';
                  $matrices['valores'][]=$fieldObject->__repeat__;
/*
                  $fieldObject->ifs->OpLogico  = $this->calculoExpresion($fieldObject->ifs->OpLogico, $fila2, $nfila, $matrizNombres, $matrizValores);
                  $fieldObject->ifs->verdadero = $this->calculoExpresion($fieldObject->ifs->verdadero, $fila2, $nfila, $matrizNombres, $matrizValores);
                  $fieldObject->ifs->falso     = $this->calculoExpresion($fieldObject->ifs->falso, $fila2, $nfila, $matrizNombres, $matrizValores);
                  */
                }
                if (is_array($matrices)) {
                    $matrizNombres = $matrices['nombres'];
                    $matrizValores = $matrices['valores'];

                }

//die();
                /*
loger($fieldObject->NombreCampo, 'if');
loger($fieldObject->ifs->verdadero, 'if');
loger($matrices, 'if');
*/
                // Fuerzo la condicion
                if ($fieldObject->ifs->OpLogico == 'true') {
                    $var = true;
                } else {
                    $operacion = $this->calculoExpresion($fieldObject->ifs->OpLogico, $fila2, $nfila, $matrizNombres, $matrizValores);
                    $var = '';
                    $val = '';

//                    $retorno1 = eval('$var = ' . $operacion . ';');
                 try {
	              $retorno1 = eval('$var = '.$operacion.';');
    	          } catch (\Error $e) {
        	      $retorno1 = false;
                      loger($this->xml.' '.$fieldObject->NombreCampo, 'eval_errors.log');
	              loger('$val = '.$operacion.';', 'eval_errors.log');
    	              loger($e->getMessage(), 'eval_errors.log');
                  } catch (\Exception $e) {
	              $retorno1 = false;
    	              loger($this->xml.' '.$fieldObject->NombreCampo, 'eval_exception.log');
        	      loger('$val = '.$operacion.';', 'eval_exception.log');
                      loger($e->getMessage(), 'eval_exception.log');
	          }


                    if ($retorno1 === false) {
                        loger($fieldObject->NombreCampo . ' - ' . ' $val = (' . $operacion
                         . ')', 'Error_if_1');
                    }
                }

                //      loger($this->xml);

                //loger($fieldObject->ifs->verdadero ,$fieldObject->NombreCampo.'.if');
                //loger($precalc ,$fieldObject->NombreCampo.'.if');
                //loger($matrices  ,$fieldObject->NombreCampo.'.if');

                if ($var) {
                    $operacion = $this->calculoExpresion($fieldObject->ifs->verdadero, $fila2, $nfila, $matrizNombres, $matrizValores);
                    if ($operacion == '')
                        $operacion = $fieldObject->ifs->verdadero;

                    //                    loger($this->xml.'__'.$fieldObject->NombreCampo.'operacion = (' . $fieldObject->ifs->verdadero . ')', 'Error_if_2');
                } else {
                    $operacion = $this->calculoExpresion($fieldObject->ifs->falso, $fila2, $nfila, $matrizNombres, $matrizValores);
                    if ($operacion == '')
                        $operacion = $fieldObject->ifs->falso;

                    //                      loger($this->xml.'__'.$fieldObject->NombreCampo.'operacion = (' . $fieldObject->ifs->falso . ')', 'Error_if_2');
                }

//                         loger($operacion, 'finif');

              try {
                  $retorno = eval('$val = '.$operacion.';');
                  //loger('$val = '.$operacion.';', 'eval_op.log');

              } catch (\Error $e) {
                  $retorno = false;
                  loger($this->xml.' '.$fieldObject->NombreCampo, 'eval_errors.log');
                  loger('$val = '.$operacion.';', 'eval_errors.log.log');
                  loger($e->getMessage(), 'eval_errors.log.log');
              } catch (\Exception $e) {
                  $retorno = false;
                  loger($this->xml.' '.$fieldObject->NombreCampo, 'eval_exception.log');
                  loger('$val = '.$operacion.';', 'eval_exception.log');
                  loger($e->getMessage(), 'eval_exception.log');
              }



            if (is_object($val)){
		loger($val, 'eval_error_obj.log');
                $val = $val();
            }
                //loger('V: '.$fieldObject->NombreCampo.' - '.$operacion, 'if');

                if ($retorno === false) {
                    loger($this->xml . '__' . $fieldObject->NombreCampo . ': $val = ' . $operacion . '', 'Error_if_2');
                }


                if (isset($fieldObject->lastValue) && $fieldObject->lastValue == "true") {
                    if ($fieldObject->ultimo != 0 && $val == '')
                        $val = $fieldObject->ultimo;
                }

                if (isset($fieldObject->aletras) && $fieldObject->aletras == true) {
                    if (is_numeric($val))
                        $val = NumeroALetras($val);
                }




                if (isset($fieldObject->retroalimenta) && $fieldObject->retroalimenta == 'true') {

                    if ($cont == 0)
                        $fieldObject->Suma = 0;
                    $fieldObject->Suma += $val;
                    $val = $fieldObject->Suma;
                    $fieldObject->updateSetters($fieldObject->Suma);

                }



            }

            if (isset($fieldObject->Expresion) && $fieldObject->expresionSql == 'true' 

              && $this->_setData != true // add this to prevent multiple querys on setData process

              ) {

                if ($fieldObject->Expresion != '')
                $val = $this->ejecutoExpresionSql($fieldObject, $this->TablaTemporal->Tabla[$nfila]);

                // this alows to use expressions to fill forms
                $fieldObject->valor = $val;
            }

            $this->TablaTemporal->Tabla[$nfila][$columna] = $val;


            /*
             * ACTIVATE IF NECESARY
             *
              if ($fieldObject->valAttribute != '') {
              foreach ($fieldObject->valAttribute as $attribID => $attrib) {

              $valAtt = $this->getCampo($attrib)->valor;
              if ($attribID == 'oculto') {
              $fieldObject->setOculto($valAtt);
              }
              $fieldObject->{$attribID} = (string) $valAtt;
              $fieldObject->Parametro[$attribID]= (string) $valAtt;
              }
              }

             */
            if (isset($this->graboasiento) && $this->graboasiento) {
                $this->addRenglonMinuta($nfila, $columna, $fieldObject);
            }

            // recorro los registros y actualizo los demas contenedores
            if (isset($this->registrosEx) && $this->registrosEx != '') {
                foreach ($this->registrosEx as $idReg => $contRegEx) {
                    $valReg = $this->TablaTemporal->Tabla[$nfila][$columna];
                    $contRegEx->campo[$contRegEx->par[$columna]] = $valReg;
                }
            }
            //if ($val != '')
            $fieldObject->setLastValue($val);

            if (isset($fieldObject->noZero) && $fieldObject->noZero == 'true' && $val == 0) {
                $val = '';
            }


            if ($val != '') {
                $this->hasValue[$fieldObject->NombreCampo] = 'true';
            }

            if (isset($fieldObject->validRow) && $val == 0) {
                $valid = 0;

                return $valid;
            }
        }

        return $valid;
    }

    public function addRenglonMinuta($nfila, $columna, $fieldObject)
    {
        $valoraCtb = $this->TablaTemporal->Tabla[$nfila][$columna];

        if (count($fieldObject->opcion) > 0 && $fieldObject->TipoDato != "check" && $fieldObject->valop == 'true') {
            $valoraCtb = $fieldObject->opcion[$valoraCtb];
            if (is_array($valoraCtb)) {
                $valoraCtb = current($valoraCtb);
            }
        }

        if ($fieldObject->ctbfecha != '') {
            $this->AsientoTemp[$fieldObject->ctbfecha]['ctbfecha'] = $valoraCtb;
        }

        if ($fieldObject->ctbreferencia != '') {
            $this->AsientoTemp[$fieldObject->ctbreferencia]['ctbreferencia'] = $valoraCtb;
        }

        if ($fieldObject->ctbcuenta != '') {
            $this->AsientoTemp[$fieldObject->ctbcuenta]['ctbcuenta'] = $valoraCtb;
        }

        if ($fieldObject->ctbregnro != '') {
            $this->AsientoTemp[$fieldObject->ctbregnro]['ctbregnro'] = $valoraCtb;
        }

        if ($fieldObject->ctbtipo != '') {
            $regdoh = $this->TablaTemporal->Tabla[$nfila][$columna];
            $this->AsientoTemp[$fieldObject->ctbtipo]['ctbtipo'] = $regdoh;
        }
        if ($fieldObject->ctbimporte != '') {
            $this->AsientoTemp[$fieldObject->ctbimporte]['ctbimporte'] = $valoraCtb;
        }
        //        loger($this->AsientoTemp);


        if ($fieldObject->ctbcentro != '') {
            $this->AsientoTemp[$fieldObject->ctbcentro]['ctbcentro'] = $valoraCtb;
        }
    }


    public function creaMinuta()
    {
          $this->Minuta = new Minuta();
          $this->Minuta->ajusteDebe = $this->ajusteDebe;
          $this->Minuta->ajusteHaber = $this->ajusteHaber;

    }

    /**
     * recalc grid
     * @param  boolean $serialize do serialize de container
     * @return [type]             [description]
     */
    public function calculointerno($serialize = true)
    {
        //$TIEMPO = processing_time();
        if (isset($this->graboasiento) && $this->graboasiento == 'true') {
          $this->creaMinuta();
        }


        // Unserialize Header
        if (isset($this->CabeceraMov)){
            foreach ($this->CabeceraMov as $nCabecera => $containerCabecera) {
                // desserializo para estar seguro que esta actualizado
                if ($serialize) {
                    $this->CabeceraMov[$nCabecera] = Histrix_XmlReader::unserializeContainer($containerCabecera);
                    $serialize = false; // prevents further unserialize
                }
            }
        }

        unset($this->Mnoms);
        unset($this->Mvals);
        if (isset($this->CabeceraMov)){
            foreach ($this->CabeceraMov as $nCabecera => $containerCabecera) {
                foreach ($containerCabecera->tablas[$containerCabecera->TablaBase]->campos as $numCamCab => $campoCabecera) {
                    $matrizNombres[] = $campoCabecera->NombreCampo;
                    $matrizValores[] = (isset($campoCabecera->valor)) ? $campoCabecera->valor : '';

                    if (isset($campoCabecera->contExterno) && isset($campoCabecera->contExterno->xml)
                            && $campoCabecera->isSelect != true) {

                        $contInterno = $campoCabecera->contExterno;
                        $contInterno = Histrix_XmlReader::unserializeContainer($campoCabecera->contExterno);

                        //if ($contInterno)
                            foreach ($contInterno->tablas[$contInterno->TablaBase]->campos as $numCamIn => $innerField) {
                                $matrizNombres[] = $innerField->NombreCampo;
                                if ($innerField->Suma != '') {
                                    $matrizValores[] = $innerField->Suma;
                                } else
                                    $matrizValores[] = $innerField->valor;
                            }
                        unset($contInterno);
                    }
                }
            }
        }
        //
        // Get data from parent Instance
        if (isset($this->parentInstance) && $this->unserializeParent != 'false') {
          unset($parentContainer);
          $parentContainer = new ContDatos("");
          $parentContainer = Histrix_XmlReader::unserializeContainer(null, $this->parentInstance);


            if ($parentContainer != '') {
                foreach ($parentContainer->tablas[$parentContainer->TablaBase]->campos as $numCamIn => $innerField) {

            // field must be declared public
                    if (isset($innerField->public) && $innerField->public == 'true') {
                        $matrizNombres[] = $innerField->NombreCampo;
                        if ($innerField->Suma != '') {
                            $matrizValores[] = $innerField->Suma;
                        } else
                            $matrizValores[] = $innerField->valor;
                    }
                }
                unset($parentContainer);

          }
         }

        if (isset($matrizNombres))
            $this->Mnoms = $matrizNombres;
        if (isset($matrizValores))
            $this->Mvals = $matrizValores;


        $cont = 0;
        if (isset($this->TablaTemporal->Tabla) ) {
            foreach ($this->TablaTemporal->Tabla as $nfila => $fila) {

                if ($fila != '')
                    $valid = $this->calculoInternoFila($fila, $nfila, $cont, true);

                if ($valid == 0) {

                    unset($this->TablaTemporal->Tabla[$nfila]);
                }
                $cont++;
                if (isset($this->graboasiento) && $this->graboasiento == 'true')
                    $MinutaParaAnexar = $this->generoAsiento($serialize);
            }
        }


        if (isset($MinutaParaAnexar))
            $this->Minuta->anexoMinuta($MinutaParaAnexar);

        // Actualizo los demas contenedores
        //$this->actualizoContenedores();
        if (isset($this->registrosEx) && $this->registrosEx != '') {
            //  $contReferente->TablaTemporal->deleteAutos();
            $del = true;
            foreach ($this->registrosEx as $idReg => $contRegEx) {
                $contRegEx->Actualizar($this, $del);
                $del = false;
            }
        }

        
        // Testing multiple inner temptables
        if (isset($this->tempTables)) {
            $this->mapTempTables();
        }
        

    }

    /**
    * Maps temporal tables
    * scan temporal tables and map data from innerTable to the new temporal tables
    */
    private function mapTempTables()
    {
        if (isset($this->tempTables ) && $this->updateTempTable != 'false') {
            foreach ($this->tempTables as $id_table => $table) {


                $table->mapData($this->TablaTemporal, $this->xml); // map data from datatable to temporal table

                // get innerTempTables to merge
                // from fields
//    loger($table, 'map');

                $fieldArray = $this->camposaMostrar();
                foreach ($fieldArray as $number => $id_field) {
                    $Field = $this->getCampo($id_field);

                    if (isset($Field->contExterno) && isset($Field->contExterno->tempTables)) {

                        $InnerContainer = Histrix_XmlReader::unserializeContainer($Field->contExterno);
                    //    $InnerContainer = $Field->contExterno;
                        $InnerContainer->calculoInterno();
  //                      
             //           $InnerContainer->mapTempTables(); // make it recursive
                        $innerTable = $InnerContainer->tempTables[$id_table];
  
                        if ($innerTable != '') {

                            $table->deleteData($InnerContainer->xml);
                            $table->appendTable($innerTable); //append Data from inner Ttempables
                        }

                    }
                }

            }
              /*
                 loger(, 'TempTable.log');
              foreach ($this->tempTables[$id_table]->Tabla as $row){
                if (trim($row['importe']) != '')
                  loger($row, 'TempTable.log');
               }
               */
        } //if
    }

    public function generoAsiento($unserialize = true)
    {
        if ($this->graboasiento == 'true') {
            $invierte = false;
            if (isset($this->CabeceraMov)) {

                foreach ($this->CabeceraMov as $ncab => $cab) {
                    if ($unserialize) // check if was previously unserialized
                        $cab = Histrix_XmlReader::unserializeContainer($cab);

                    $campos = $cab->camposaMostrar();
                }
                foreach ($campos as $ncam => $campo) {
                    $objCampo = $cab->getCampo($campo);
                    $valoraCtb = $objCampo->getValor();

                    // Chequeo si tiene que invertirse el asiento
                    if ($objCampo->ctbinvierte == 'true') {

                        if ($valoraCtb === 'true' || $objCampo->valor === true || 
                          $valoraCtb === '1' || $valoraCtb === 1) {
                            $invierte = true;
                        }
                    }
                    if (count($objCampo->opcion) > 0 && $objCampo->TipoDato != "check" && $objCampo->valoropcion == 'true') {

                        $valoraCtb = $objCampo->opcion[$objCampo->getValor()];
                        if (is_array($valoraCtb)) {
                            $valoraCtb = current($valoraCtb);
                        }

                    }

                    if ($objCampo->ctbcuenta != '') {
                        $this->AsientoTemp[$objCampo->ctbcuenta]['ctbcuenta'] = $valoraCtb;
                    }

                    if ($objCampo->ctbcentro != '') {
                        $this->AsientoTemp[$objCampo->ctbcentro]['ctbcentro'] = $valoraCtb;
                    }


                    if ($objCampo->ctbreferencia != '') {
                        $this->AsientoTemp[$objCampo->ctbreferencia]['ctbreferencia'] = $valoraCtb;
                    }

                    if ($objCampo->ctbfecha != '') {
                        $this->AsientoTemp[$objCampo->ctbfecha]['ctbfecha'] = $objCampo->valor;
                    }

                    if ($objCampo->ctbsubsistema != '') {
                        $this->AsientoTemp[$objCampo->ctbsubsistema]['ctbsubsistema'] = $objCampo->valor;
                    }

                    if ($objCampo->ctbregnro != '') {
                        $this->AsientoTemp[$objCampo->ctbregnro]['ctbregnro'] = $valoraCtb;
                    }

                    if ($objCampo->ctbtipo != '') {
                        $this->AsientoTemp[$objCampo->ctbtipo]['ctbtipo'] = $valoraCtb;
                    }


                    /* Si el campo tiene un contenedor interno con Minuta
                     * Tengo que tomar ese asiento e incorporarlo
                     */
                    if ($objCampo->contExterno != '' && $objCampo->contExterno->xml != ''
                            && $objCampo->isSelect != true) {

                        if ($objCampo->contExterno->graboasiento == 'true' || (isset($contActual) && $contActual->Minuta)) {
                            $contActual = new ContDatos("");
                            $contActual = Histrix_XmlReader::unserializeContainer($objCampo->contExterno);
                        }
                    }
                }
            }

            if (isset($this->AsientoTemp)) {
                foreach ($this->AsientoTemp as $Nasi => $asi) {
                    if (isset($asi['ctbfecha'])) {
                        $this->Minuta->fecha = $asi['ctbfecha'];
                    }

                    if (isset($asi['ctbsubsistema'])) {
                        $this->Minuta->subsistema_id = $asi['ctbsubsistema'];
                    }


                    if (isset($asi['ctbreferencia'])) {
                        $this->Minuta->referencia = $asi['ctbreferencia'];
                    }

                    if (isset($asi['ctbregnro'])) {
                        $this->Minuta->ctbregnro = $asi['ctbregnro'];
                    }

                    if (isset($asi['ctbcentro'])) {
                        $this->Minuta->ctbcentro = $asi['ctbcentro'];
                    }


                    $this->Minuta->addRenglon($asi['ctbcuenta'], $asi['ctbtipo'], $asi['ctbimporte'], '', $asi['ctbcentro']);
                }
            }
        }

        if ($invierte) {
            $this->Minuta->invierto();
        }


        if (isset($contActual) && $contActual->Minuta) {
            // supongo que el de caja y el recibo
            return $contActual->Minuta;
        }
    }

    public function Insert()
    {
        /* si es un ingreso se usa la tabla interna */
        $autoinc = false;
        if ($this->tipoAbm == 'ing' || $this->tipoAbm == 'grid') {
            $arrayin = $this->getInsertTemporal();

            $autoinc = $this->TablaTemporal->insert($arrayin, $this->autoUpdateRow);

            $this->calculointerno();

        } else {
            $str = trim($this->getInsert());

            // LDAP UPDATE
            if (isset($this->ldap) && $this->ldap == 'true') {
                $ldap = new ldapConnector();
            }


            // mmmm  
            if ($this->CuerpoMov != '') {
              $this->TablaTemporal->emptyTable();
              $this->TablaTemporal->insert($this->getInsertTemporal(), $this->autoUpdateRow);
            }

            // generate aditional movements BEFORE MAIN INSERT
            // test with current xml's
            $movementsResponse = $this->processMovements('insert', 'before');

            if ($str != '') {
                $autoinc = updateSQL($str, 'insert', $this->xml);

                // rollback support
                if ($autoinc === -1) {
                    return -1;
                }

            }

            // Actualizo numeradores si es que hay
            foreach ($this->tablas[$this->TablaBase]->campos as $numCampoParam => $campoConParametros) {
                /* Ldap update */
                if (isset($campoConParametros->ldifName) && $campoConParametros->ldifName != '' && isset($ldap)) {

                    $valor = $campoConParametros->getNuevoValor();
                    $keys = explode(',', $campoConParametros->ldifName);
                    foreach ($keys as $index => $key) {
                        $ldap->addData($key, $valor);
                        if ($campoConParametros->ldifKey == "true") {
                            $ldap->addKeyNew($key, $valor);
                            $ldap->addKeyOld($key, $campoConParametros->getValor());
                        }
                    }
                }

                if (isset($campoConParametros->ContParametro)) {
                    $campoParRet = $campoConParametros->ContParametro->getCampo($campoConParametros->ContParametro->retorna);
                    if (isset($campoParRet->onComplete)) {
                        $nuevoValor = $campoParRet->onComplete;
                        $this->setParametros($campoConParametros, $nuevoValor, false);
                    }
                }
            }

                   /*
                // process inners
                foreach ($this->tablas[$this->TablaBase]->campos as $numfield => $field) {
                    if (is_object($field->contExterno)){
                      if ($field->contExterno->CuerpoMov != '')                                     
                        $field->contExterno->grabarRegistros();
                    }
                
                }

                     */
            // Log action into log table
            if (isset($this->log) && $this->log == "true") {
                // new log implementation
                $this->lastLogcount += 1;

                if ($this->logReference != '') {
                    $references = explode(',', $this->logReference);
                    $this->wherelog['__ref__'] = '';
                    foreach ($references as $reference) {
                        $this->wherelog['__ref__'] .= $this->getCampo($reference)->valor . ' ';
                    }

                }
                $titlog = ($this->titulo_div != '')?$this->titulo_div:$this->titulo;
                $loger = new Histrix_Loger($this->xml, $titlog, $this->wherelog, $this->updatelog);
                $loger->dir = $this->dirxml;
                $loger->log('insert');
            }

            // LDAP INSERT SUPPORT
            if (isset($ldap)) {
                $ldap->insert();
                $ldap->close();
            }

            // process aditional movements
            $movementsResponse = $this->processMovements('insert');

            if ($movementsResponse === false) {
              return false;
            }

        }

        return $autoinc;
    }

    //
    private function processMovements($triggerEvent, $before = 'false')
    {



            $updateResponse = 0;
            // inserts adicionales <movimientos>
            // REFACTOR WITH SAME CODE IN UPDATE AND DELETE
            if ($this->CuerpoMov != '') {
                foreach ($this->CuerpoMov as $numMoim => $relationContainer) {
            $relationContainer->_skip = true;

                    if ( isset($relationContainer->evento) )
                      if( $relationContainer->evento != $triggerEvent) continue;

                    // check if this xml should be executed before or after main event.
                    if (($relationContainer->before == 'true' && $before == 'false') ||
                    ($relationContainer->before != 'true' && $before != 'false')) continue;
        
                $relationContainer->_skip = false;
        
                    $updateResponse = $this->GrabarRegistros();
                }
            }

            return $updateResponse;
    }

    public function getInsertTemporal()
    {
        foreach ($this->tablas[$this->TablaBase]->campos as $miNro => $fieldItem) {

            if ($fieldItem->esOculto())
                continue;
            $strcampo = $fieldItem->NombreCampo;
            $valorcampo = $fieldItem->getNuevoValor();

            $arrayin[$miNro] = replaceSlashes($valorcampo);
            if ($fieldItem->innerTable != '') {
                $arrayin[$miNro] = $fieldItem->innerTable->datos();
            }
        }

        return $arrayin;
    }

    public function getCustomScript()
    {
        $script = $this->customScript;

        $script = str_replace('[__instance__]', $this->getInstance(), $script);
        $script = str_replace('__instance__', $this->getInstance(), $script);

        return $script;
    }


    public function getInsert()
    {
        // $this->wherelog='';

         if (isset($this->customSql))
            return $this->customSql;

        $pares = '';
        $campos = '';

        $tablains = $this->tablas[$this->TablaBase]->getNombre();
        $action = 'INSERT';
        $tipoInsert = (isset($this->tipoInsert)) ? $this->tipoInsert : '';

        if (isset($this->replaceInto) && $this->replaceInto == 'true' ||
            (isset($this->tipoInsert) && $this->tipoInsert == 'REPLACE')) {
            $action = 'REPLACE';
            $tipoInsert = '';
        }
        // Get Database
        $database = (isset($this->database)) ? $this->database . '.' : '';

        $str = $action . ' ' . $tipoInsert . ' into ' . $database . $tablains;
        if (isset($this->tablas[$this->TablaBase]->campos))
            foreach ($this->tablas[$this->TablaBase]->campos as $miNro => $fieldItem) {

                if ((isset($fieldItem->autoinc) && $fieldItem->autoinc ) ||
                        (isset($fieldItem->local) && $fieldItem->local ) ||
                        (isset($fieldItem->Expresion) && $fieldItem->Expresion != '' ) || (isset($fieldItem->tabla) && $fieldItem->tabla != '' && $fieldItem->tabla != $this->TablaBase)) {
                    continue;
                }

                $strcampo = $fieldItem->NombreCampo;

                // Replace fiedlName for calculated fieldName
                if (isset($fieldItem->idName) && $fieldItem->idName != '') {
                    $strcampo = $this->getCampo($fieldItem->idName)->valor;
                }

                $valorcampo = $fieldItem->getNuevoValor();

                if ($valorcampo == '' && $fieldItem->valorOrig == '_NRO_MIN') {
                    $valorcampo = $fieldItem->valorOrig;
                }

                if ($strcampo == '')
                    continue;

                if ($valorcampo == '' && $fieldItem->required =="true" &&  $this->tipoAbm  == 'insert') {
                    return;
                }

                $tipodatocampo = $fieldItem->TipoDato;
                if ($valorcampo == '' && $tipodatocampo == 'date')
                    continue;

                if (isset($fieldItem->FType) && $fieldItem->FType == 'password' && $fieldItem->FEncode != 'plain') {

                    $valorcampo = md5($valorcampo); // TODO define a global encription function
                }

                //  $valorcampo = addslashes($valorcampo);
                $strcampovalor = '';

                // Not_So_Magic_Quotes Value
                $quotedValue = Types::getQuotedValue($valorcampo, $tipodatocampo, 'xsd:integer');
                if ($quotedValue == '')
                    $quotedValue = 0;
                $strcampovalor = $quotedValue;
                $pares[$strcampo] = $strcampovalor;

                $campos .= $strcampo;

                if (isset($this->log) && $this->log == 'true') {
                    if ($strcampovalor != '' && $strcampovalor != "''") {

                        if ($fieldItem->opcion != '') {
                            $strcampovalor = $fieldItem->opcion[$valorcampo];
                            if (is_array($strcampovalor))
                                $strcampovalor = current($strcampovalor);
                        }
                        if ($fieldItem->TipoDato == 'check')
                            $strcampovalor = ($strcampovalor == 1) ? 'Si' : 'No';

                        $et = ($fieldItem->Etiqueta != '') ? $fieldItem->Etiqueta : $strcampo;
                        if ($fieldItem->log != 'false')
                            $this->wherelog[$et] = $strcampovalor;
                    }
                }
            }
        $first = true;
        $valcampos = '';
        $campos = '';
        if(is_array($pares))
        foreach ($pares as $nompar => $valorpar) {

            $customSelect = (isset($this->customSelect)) ? $this->customSelect : '';
            if ($customSelect == '' && $valorpar == '')
                continue;

            if ($first == false)
                $coma = ' , ';
            else
                $coma = '';
            $campos .= $coma . '`' . $nompar . '`';
            $valcampos .= $coma . $valorpar;

            $sqlArr[] = "`$nompar` = $valorpar";
            $first = false;
        }

        if (isset($this->sqlhardcoded) && $this->sqlhardcoded == "true") {
            // esto esta MAL tiene que estar disparado por otro trigger
            // ver DONDE se hace para modificar el programa
            loger('Arreglar este xml' . $this->xml . ' que usa sqlharcoded de forma no deseada', 'to_fix.log');
            $str .= ' ' . $valcampos;

            return $str;
        }

        $str .= ' ( ' . $campos . ' ) ';

        if (isset($this->customSelect)) {
            $str .= $this->customSelect;
        } else {
            $str .= ' VALUES ';
            $str .= ' ( ' . $valcampos . ' ) ';
        }

        if (isset($this->onDuplicateKey) && $this->onDuplicateKey == "true") {
            $str .= ' on duplicate key update ';
            $this->getUpdate();
            $str .= $this->UpdateSets;

            return $str;
        }

         if (isset($this->customSql))
            return $this->customSql;

       if (isset($this->customSelect)) {
           return $str;
       }

    if (is_array($sqlArr))
        return "$action $tipoInsert INTO $database$tablains SET " . implode($sqlArr,',');

        // SORRY LUIS
        // return $str;
    }

    public function Update($nrofila = '', $retorna = false)
    {
        $updated = '';
        if ($this->tipoAbm == 'ing' || $this->tipoAbm == 'grid') {

            $valores = $this->getInsertTemporal();
            $this->TablaTemporal->updateFila($nrofila, $valores);

        } else {

            if ($this->transaccional == 'true')
                _begin_transaction();

            // generate aditional movements BEFORE MAIN INSERT
            // test with current xml's

      // mmmm  
            if ($this->CuerpoMov != '') {

        $this->TablaTemporal->emptyTable();
              $this->TablaTemporal->insert($this->getUpdateTemporal(), $this->autoUpdateRow);
      }

            $movementsResponse = $this->processMovements('update', 'before');

            if ($movementsResponse === -1)
                return -1;

            $checkDupOnUpdate = (isset($this->checkDupOnUpdate)) ? $this->checkDupOnUpdate : null;
            $str = $this->getUpdate($checkDupOnUpdate);

              $updated = 0;


            if (trim($str) != '' && $retorna == false) {
                $updated = updateSQL($str, null, $this->xml);
                // rollback support
                  if ($updated === -1) {
                    return -1;
                  }
                }

            // LDAP UPDATE
        // TODO: REMOVE LDAP FROM HERE
            if ($this->ldap == 'true') {
                $ldap = new ldapConnector();

                foreach ($this->tablas[$this->TablaBase]->campos as $numfield => $field) {
                    /* Ldap update */
                    if ($field->ldifName != '' && $this->ldap == 'true') {

                        $valor = $field->getNuevoValor();
                        $keys = explode(',', $field->ldifName);
                        foreach ($keys as $index => $key) {
                            $ldap->addData($key, $valor);
                            if ($field->ldifKey == "true") {
                                $ldap->addKeyNew($key, $valor); // Key is old Value
                                $ldap->addKeyOld($key, $field->getValor()); // Key is old Value
                            }
                        }
                    }
                }
                $ldap->update();
            }
  
  /*
                // process inners
                foreach ($this->tablas[$this->TablaBase]->campos as $numfield => $field) {
                    if (is_object($field->contExterno)){
                      if ($field->contExterno->CuerpoMov != '')                                     
                        $field->contExterno->grabarRegistros();
                    }
                
                }
    */


            if ($this->log == "true") {
                // New log implementation
                $this->lastLogcount = 0;
                $titlog = ($this->titulo_div != '')?$this->titulo_div:$this->titulo;
                $loger = new Histrix_Loger($this->xml, $titlog, $this->wherelog, $this->updatelog);
                $loger->dir = $this->dirxml;
                $loger->log('update');
                $this->lastLogcount = $loger->logcount;
                unset($this->updatelog);
            }


            if ($this->transaccional == 'true')
                _end_transaction();

            $valores = $this->getInsertTemporal();
            $claves = $this->getUpdateTemporal();
            $this->TablaTemporal->update($claves, $valores);


            // process aditional movements
            if ($updated != 0) {
                $movementsResponse = $this->processMovements('update');

                if ($movementsResponse === -1) return -1;
            }


        }
        // de que me sirve tener el calculo aca??? pregunto yo
        // lo comento a ver que carajo pasa...

        $this->calculointerno(); // Si lo saco se rompe el ingreso de fact de proov.
        //if ($updated != '')
        return $updated;

        //return $str;
    }

    public function getUpdateTemporal()
    {
        /*
          $hayClave = false;
          // controlo si existe al menos un campo clave para verificar
          foreach ($this->tablas[$this->TablaBase]->campos as $miNro => $fieldItem) {
          if (($fieldItem->Parametro['esclave']))
          $hayClave = true;
          }
         */
        if (isset($this->tablas[$this->TablaBase]->campos))
            foreach ($this->tablas[$this->TablaBase]->campos as $miNro => $fieldItem) {

                if ($fieldItem->Expresion != '')
                    continue;
                $valorcampo = $fieldItem->getValor();

                if ($valorcampo == '')
                    continue;

                $strcampo = $fieldItem->NombreCampo;

                if (strlen($strcampo) > 0) {
                    $campos[$strcampo] = $valorcampo;
                }
            }

        return $campos;
    }

    public function getUpdate($nodup = false, $force = false)
    {
         if (isset($this->customSql))
            return $this->customSql;

        $tablaupd = $this->tablas[$this->TablaBase]->getNombre();
        $database = (isset($this->database)) ? $this->database . '.' : '';
        $str = 'UPDATE ' . $database . $tablaupd;
        // Add Joins
        $strJoins = '';
        if ($this->tipoAbm == 'update')
            $strJoins = $this->getJoins();
        $str .= $strJoins;
        $str .= ' set ';
        /* valores del form */

        $first = true;
        $first2 = true;
        $contad = 0;
        $hacerupdate = false;
        $campos = '';
        if (isset($this->tablas[$this->TablaBase]->campos))
            foreach ($this->tablas[$this->TablaBase]->campos as $miNro => $fieldItem) {

                if ($this->tipoAbm == 'update') {
                    if ($fieldItem->esClave == true && $fieldItem->Parametro['esclave'] != 'false') {

                        $strcampowhere = $fieldItem->NombreCampo;

                        if ($strJoins != '') {
                            if ($fieldItem->TablaPadre != '')
                                $tabla = $fieldItem->TablaPadre;
                            if ($fieldItem->alias != '')
                                $tabla = $fieldItem->alias;
                            if ($tabla != '')
                                $tabla .= '.';

                            $strcampowhere = $tabla . $fieldItem->NombreCampo;
                        }

                        $valorcampowhere = $fieldItem->getNuevoValor();

                        if (strlen($strcampowhere) > 0) {

                            if ($first2 == false)
                                $camposwhere .= ' and ';
                            else
                                $first2 = false;

                            // Not_So_Magic_Quotes Value
                            $quotedValuewhere = Types :: getQuotedValue($valorcampowhere, $fieldItem->TipoDato, 'xsd:integer');
                            if ($quotedValuewhere == '')
                                $quotedValuewhere = 0;


                            $operador = (isset($fieldItem->operador))?' '.$fieldItem->operador.' ':' = ';
                            $strcampowhere .= $operador . $quotedValuewhere;

                            $camposwhere .= $strcampowhere;
                            $contad++;
                        }
                    }
                }

                if ($fieldItem->local)
                    continue; // excluyo campos locales
         if ($fieldItem->Expresion != '')
                    continue; // excluyo expresiones
         if ($fieldItem->Oculto == true)
                    continue; // excluyo expresiones

                if ($fieldItem->tabla != '' && $fieldItem->tabla != $tabla)
                    continue; // excluyo updates sobre campos de otra tabla

                    $strcampo = $fieldItem->NombreCampo;

                if ($strJoins != '') {
                    $tabla = '';

                    if ($this->TablaBase != '')
                        $tabla = $this->TablaBase;

                    if ($fieldItem->TablaPadre != '')
                        $tabla = $fieldItem->TablaPadre;
                    if ($fieldItem->alias != '')
                        $tabla = $fieldItem->alias;
                    if ($tabla != '')
                        $tabla .= '.';
                    $strcampo = $tabla . $fieldItem->NombreCampo;
                }

                // ver que metodo corresponde, probar

                $oldvalorcampo = $fieldItem->getValor();
                $valorcampo    = $fieldItem->getNuevoValor();

                // si los datos del where y del set coinciden no hago set
                if ($nodup == false) {
                    if ($this->tipoAbm != 'ficha'){
                        if ($fieldItem->valauto == ''){
                            if ($oldvalorcampo == $valorcampo && $fieldItem->forceUpdate != 'true'){
                                continue;
                            }
                        }
                    }
                }

                $hacerupdate = true;
                //$force = true;
                if (strlen($strcampo) > 0 || $force == true) {

                    if ($first == false)
                        $strcampo = ' , ' . $strcampo;
                    else
                        $first = false;

                    if ($fieldItem->FType == 'password' && $fieldItem->FEncode != 'plain') {
                        if ($oldvalorcampo != $valorcampo)
                            $valorcampo = md5($valorcampo);
                    }

                    // Not_So_Magic_Quotes Value
                    $quotedValue = Types :: getQuotedValue($valorcampo, $fieldItem->TipoDato, 'xsd:integer');

                    if ($quotedValue == '')
                        $quotedValue = 0;


                    $strcampo .= ' = ' . $quotedValue;

                    if ($this->log == 'true') {
                        $et = ($fieldItem->Etiqueta != '') ? $fieldItem->Etiqueta : $strcampo;
            if ($fieldItem->alertStyle != '') {
                $et = '<span style=\"'.$fieldItem->alertStyle.'\">'.$et.'</span>';
            }
                        $valueTxt = $valorcampo;
                        if ($fieldItem->opcion != '') {
                            $valueTxt = $fieldItem->opcion[$valorcampo];
                            if (is_array($valueTxt))
                                $valueTxt = current($valueTxt);
                        }

                        if ($fieldItem->TipoDato == 'check')
                            $valueTxt = ($valorcampo == 1) ? $this->i18n['yes'] : $this->i18n['no'];

                        // this prevents logging where integer is set to empty.
                        $tipo = Types::getTypeXSD($dataType, 'xsd:integer');
                        if ($tipo == 'xsd:integer' || $tipo == 'xsd:decimal') {
                            if ($valorcampo == '')
                                $valorcampo = 0;
                        }

                        if ($valorcampo != $oldvalorcampo || isset($fieldItem->forceUpdate) && $fieldItem->forceUpdate == 'true') {

                            $this->updatelog[$et] = array($oldvalorcampo, $valorcampo, $valueTxt, time() , $this->_instance );
                        }
                        $valueTxt = '';
                    }
                }
                $campos .= $strcampo;
            }
        $this->UpdateSets = $campos;
        $str .= ' ' . $campos;
        $campos = '';
        /* condiciones para ubicar el campo actual */

        $str .= ' where ';

        if ($this->tipoAbm == 'update') {
            if ($contad == 0)
                $camposwhere = '';
            $campos = $camposwhere;
        } else
            $campos = $this->where_actual();

        $str .= ' ' . $campos;

        $limit = 1;

        if ($this->noLimit == 'true')
            $this->limit = 0;
        //$str .= ' LIMIT 1';

        $limit = ($this->limit != '') ? $this->limit : 1;

        if ($limit > 0)
            $str .= ' LIMIT ' . $limit;

        if ($campos == '' || $hacerupdate == false)
            $str = '';

        return $str;
    }

    /**
     * Cadena de condiciones para un where SQL que ubica el registro actual
     */
    public function where_actual()
    {
        $this->wherelog = '';
        $first = true;

        $hayClave = false;
        // controlo si existe al menos un campo clave para verificar
        foreach ($this->tablas[$this->TablaBase]->campos as $miNro => $fieldItem) {

            if ($fieldItem->Parametro['esclave'] == 'true' || $fieldItem->esClave == true || $fieldItem->Parametro['esClave'] == 'true') {
                $hayClave = true;
            }
        }

        $contad = 0;
        $campos = '';
        if (isset($this->tablas[$this->TablaBase]->campos))
            foreach ($this->tablas[$this->TablaBase]->campos as $miNro => $fieldItem) {
                // excluyo los campos locales
                if ($fieldItem->local)
                    continue;
                if ($fieldItem->Expresion != '')
                    continue;
            if ($hayClave) {
              if ($fieldItem->Parametro['esclave'] == 'true' || $fieldItem->esClave == true || $fieldItem->Parametro['esClave'] == 'true') {
//                echo $fieldItem->NombreCampo;

              } else {
             continue;
              }
        }

/*                if ((!($fieldItem->esClave) && $hayClave) || ($hayClave && $fieldItem->esClave == 'false')) {
                echo $fieldItem->NombreCampo;
                    continue;
                }
*/
                $valorcampo = $fieldItem->getValor();
                $Newvalorcampo = $fieldItem->getNuevoValor();

                if ($valorcampo == '' && $Newvalorcampo != '') {
                    if (!($fieldItem->esClave))
                        continue;
                }

                $strcampo = '';
                $strcampo = $fieldItem->NombreCampo;

                $tabla = '';
                if ($this->TablaBase != '')
                    $tabla = $this->TablaBase;
                if ($fieldItem->TablaPadre != '')
                    $tabla = $fieldItem->TablaPadre;
                if ($fieldItem->alias != '')
                    $tabla = $fieldItem->alias;
                if ($tabla != '') {
                    $tabla .= '.';
                    $strcampo = $tabla . $fieldItem->NombreCampo;
                }

                if (strlen($strcampo) > 0) {

                    if ($first == false)
                        $campos .= ' and ';
                    else
                        $first = false;

                    // Not_So_Magic_Quotes Value
                    $quotedValue = Types::getQuotedValue($valorcampo, $fieldItem->TipoDato, 'xsd:integer');
                    if ($quotedValue == '')
                        $quotedValue = 0;
                    $strcampo .= ' = ' . $quotedValue;

                    if ($this->log == 'true') {
                        $et = ($fieldItem->Etiqueta != '') ? $fieldItem->Etiqueta : $strcampo;
                        if ($this->logReference != '') {
                            $references = explode(',', $this->logReference);
                            foreach ($references as $reference)
                                $this->wherelog['__ref__'] .= $this->getCampo($reference)->valor . ' ';
                        }
                        // else {
                        if ($fieldItem->log != 'false')
                            $this->wherelog[$et] = $quotedValue;
                        // }
                    }

                    $campos .= $strcampo;
                    $contad++;
                }
            }

        if ($contad == 0)
            $campos = '';

        return $campos;
    }

    /**
     *  generate SQL delete string
     * @return string sql delete string
     */
    public function getDelete()
    {
        $tabladel = $this->tablas[$this->TablaBase]->getNombre();
        $str = 'DELETE from ' . $tabladel;
        $str .= ' where ';
        /* condiciones para ubicar el campo actual */

        $campos = $this->where_actual();
        $str .= ' ' . $campos;
        $limit = ($this->limit != '') ? $this->limit : 1;

        if ($limit > 0)
            $str .= ' LIMIT ' . $limit;

        if ($campos == '')
            $str = '';

        return $str;
    }

    /**
     * Generate Delete Statement
     * if $row is set then deleted current row in Temporal Table
     * @param integer $row Row Number of Temporal Table
     */
    public function Delete($row = '')
    {
        if ($row >= '' && ($this->tipoAbm == 'ing' || $this->tipoAbm == 'grid')) {
            $this->TablaTemporal->deleteRow($row);

            $this->calculointerno();
        } else {

            // generate aditional movements BEFORE MAIN INSERT
            // test with current xml's
            $movementsResponse = $this->processMovements('delete', 'before');

            $str = $this->getDelete();

            // CONTROLAR ACA LOS REGISTROS A SER BORRADOS con un select previo.
            if (trim($str) != '') {
                $updated = updateSQL($str, null, $this->xml);

                // rollback support
                if ($updated === -1)
                    return false;
            }

            // LDAP UPDATE
            if ($this->ldap == 'true') {
                $ldap = new ldapConnector();

                foreach ($this->tablas[$this->TablaBase]->campos as $numfield => $field) {
                    /* Ldap update */



                    if ($field->ldifName != '' && $this->ldap == 'true') {

                        $valor = $field->getValor();
                        $valorNew = $field->getNuevoValor();
                        $keys = explode(',', $field->ldifName);
                        foreach ($keys as $index => $key) {
                            $ldap->addData($key, $valor);
                            if ($field->ldifKey == "true") {
                                $ldap->addKeyNew($key, $valorNew);
                                $ldap->addKeyOld($key, $valor); // Key is old Value
                            }
                        }
                    }
                }
                $ldap->delete();
            }

/*
  removed temporay
  check if this is necessary
                // process inners
                foreach ($this->tablas[$this->TablaBase]->campos as $numfield => $field) {
                    if (is_object($field->contExterno)){
                      if ($field->contExterno->CuerpoMov != '')                                     
                        $field->contExterno->grabarRegistros();
                    }
                
                }
  */

            if ($this->log == "true") {
                // New loger implementation
                $this->lastLogcount = 1;
                $titlog = ($this->titulo_div != '')?$this->titulo_div:$this->titulo;
                $loger = new Histrix_Loger($this->xml, $titlog, $this->wherelog, $this->updatelog);
                $loger->dir = $this->dirxml;
                $loger->log('delete');
            }

            // process aditional movements
            $movementsResponse = $this->processMovements('delete');
            if ($movementsResponse === -1) return -1;

        }
    }

    /* funcion para obtener en el caso de los arboles el campo padre a relacionar */

    public function getPadre()
    {
        foreach ($this->tablas as $numT => $tabObj)
            foreach ($tabObj->campos as $numC => $campoOb)
                if ($campoOb->Arbol == 'padre')
                    return $campoOb->NombreCampo;
    }

    public function setEtiqueta($nombre, $Etiqueta, $table = '')
    {
        $this->getCampo($nombre)->Etiqueta = $Etiqueta;
        $this->tablas[$this->TablaBase]->etiquetas_reverse[$Etiqueta] = $nombre;
    }

    public function setSize($nombre, $size, $table = '')
    {
        $field = $this->getCampo($nombre);
        $field->Size = $size;
        $field->size = $size;
    }
/*
    public function setFormato($nombre, $Formato)
    {
        $this->getCampo($nombre)->Formato = $Formato;
    }
*/
    public function setTipo($nombre, $Tipo, $table = '', $aletras = '')
    {
        $this->getCampo($nombre)->setTipoDato($Tipo);
        if ($aletras != '')
            $this->getCampo($nombre)->aletras = $aletras;
    }

    public function setTipoAbm($TipoAbm)
    {
        $this->tipoAbm = $TipoAbm;
    }

    public function setAyuda($nombre, $hlp, $table = '')
    {
        $this->getCampo($nombre)->setAyuda($hlp);
    }



    public function setDetalle($nombre, $det, $table = '')
    {
        $this->getCampo($nombre)->setDetalle($det);
    }

    /**
     * Oculta el campo en los resultados
     */
    public function setOculto($nombre, $val, $table = '')
    {
        $this->getCampo($nombre)->setOculto($val);
    }

    public function setArbol($nombre, $val, $table = '')
    {
        $this->getCampo($nombre)->setArbol($val);
    }

    public function ocultar($campo_a_ocultar, $valor = true)
    {
        $objCampo = $this->getCampo($campo_a_ocultar);
        if ($objCampo)
            $objCampo->Oculto = $valor;
    }

    public function getAyuda($nombre, $table = '')
    {
        $this->getCampo($nombre)->getAyuda();
    }


    /**
     *
     * @param  <type> $attr
     * @return <type> Get field by Attribute
     */
    public function getFieldByAttribute($attr)
    {
        $table = $this->TablaBase;

        foreach ($this->tablas[$table]->campos as $name => $field) {
            $return = true;
            $returnField = null;
            foreach ($attr as $key => $value) {

                if (isset($field->{$key}) && $field->{$key} == $value) {

                    $returnField = $field;
                } else
                    $return = false;
            }
            if ($return == true && $returnField != null) {
                return $returnField;
            }
        }

        return false;
    }

    public function getCampo($nombre, $table = '')
    {
        if ($table == '')
            $table = $this->TablaBase;

        if ($this->tablas[$table]->getCampo($nombre, $table) == '') {

            if (isset($this->Joins))
                foreach ($this->Joins as $nro => $join) {
                    $ret = $join->getCampo($nombre);
                    if ($ret != '')
                        return $ret;
                }
        } else {
            return $this->tablas[$table]->getCampo($nombre);
        }

        return false;
    }

    public function & getCampoRef($nombre, $table = '') {
        if ($table == '')
            $table = $this->TablaBase;

        if ($this->tablas[$table]->getCampoRef($nombre, $table) == '') {

            if (isset($this->Joins))
                foreach ($this->Joins as $nro => $join) {
                    $ret = $join->getCampoRef($nombre);
                    if ($ret != '')
                        return $ret;
                }
        } else {
            return $this->tablas[$table]->getCampoRef($nombre);
        }
        $null = null;

        return $null;
    }

    /**
     * Get Value of field of inner container
     * @param string field name that has an inner container
     * @param string inner container field name to value of
     */
    private function getInnerValue($nomCampoCont, $nomCampoInt)
    {
        $contInterno = $this->getCampo($nomCampoCont)->contExterno;
        if ($contInterno->xml != '') {
            $contInterno = Histrix_XmlReader::unserializeContainer($contInterno);

            if ($contInterno) {
                //loger('calculoint grabaregs2', 'updatessql.log');
                $contInterno->calculointerno();
                $campoInt = $contInterno->getCampo($nomCampoInt);
                if ($campoInt->suma == 'true') {
                    $valor = $campoInt->Suma;
                } else
                    $valor = $campoInt->ultimo;
            }
        }

        return $valor;
    }

    public function getTitulo()
    {
        return $this->titulo;
    }

    public function setTitulo($tit)
    {
        $this->titulo = $tit;
    }

    /**
     * Notify Action to others
     * @param string $accionDes
     */
    public function Notify($action)
    {
        //echo 'NOTIFY!!!!!!!!'.$this->lastLogcount ;
        if ($this->lastLogcount != 0) {

            $notificationParameters = null;

            if (isset($this->notificationText))
                $notificationParameters['text'] = $this->notificationText;

            $notificationParameters['link'] = (isset($this->notificationLink)) ? $this->notificationLink : $this->xml;
            $notificationParameters['dir'] = (isset($this->notificationDir)) ? $this->notificationDir : $this->xmldir;

            $notificationMsg = new NotificationMessage($action, $this->getTitulo(), $notificationParameters);
            $notificationMsg->save();
        } else {

        //    loger('no ejecuto la notificacion en' . $this->xml . '_' . $action);
        }
    }


// TO MOVE ELSEWHERE
// TODO MOVE TO EXTERNAL CLASS

    public function getHelpLink()
    {
            $dir = (trim($this->dirxml) != '')?$this->dirxml:$this->subdir;

            $dir = '/'.$dir;

            $dir = str_replace('//', '/', $dir);
            $dirlink = explode('/', $dir);

            $helpId  = $dirlink[1].':'.$this->getTitulo();
            $helpId  = str_replace(array(' ' ,'/', '.'), '_', $helpId );
            $helpId  = 'http://www.estudiogenus.com/manual/?title=' . $helpId;

            return $helpId;
    }


    /**
     * Filter current Search based on posted data
     * It filters ONLY the SECOND available Field 
     * @param  string $cadena text to search
     * @return none
     */
    public function filterString($cadena)
    {
        $i = 0;
        $specialSearchStrings = '';
        if ($cadena != '') {

            // HACK PARA ARREGLAR LA BUSQUEDA CON Ñ o ñ que no funciona bien
            // REMOVED FIELD MUST BE DEFINED AS utf8_spanish_ci TO WORK PROPERLY
            /////////////////////////////////////////////////////////////////////
            /*
            if (strpos($cadena, 'ñ') !== false || strpos($cadena, 'Ñ') !== false) {
                $specialSearchStrings[1] = strtoupper_es($cadena);
                $specialSearchStrings[2] = strtolower_es($cadena);
            }
             */
            // ERROR: no me gusta la implementacion MEJORAR
            // Agregar soporte para buscar por el campo clave Y la descripcion
            foreach ($this->tablas[$this->TablaBase]->campos as $miNro => $objCampoF) {

                $nomcamp = $objCampoF->NombreCampo;
                if (($objCampoF->esOculto()))
                    continue;

                $i++;
                if ($i == 2){

                    $operador     = ' like ';
                    $reemplazo    = 'reemplazo';
                    $searchString = explode(' ', $cadena);
                    $logicop      = ' and ';


                    $oldsearch = $this->oldSearch;
                    if ($oldsearch == 'true') {
                        $searchString[] = $cadena;
                    }

                    $this->lastSearchString = $searchString;


                    if ($specialSearchStrings != '') {
                        foreach ($specialSearchStrings as $grupo => $cadena) {
                            $searchString = explode(' ', $cadena);
                            foreach ($searchString as $num => $cadena) {
                                $valor = "%" . $cadena . "%";
                                if (Types::getTypeXSD($objCampoF->TipoDato) == 'xsd:string') {
                                    $this->addCondicion($nomcamp, $operador, "'" . $valor . "'", $logicop, $reemplazo, null, $grupo);
                                    $reemplazo = '';
                                }
                            }
                        }
                    } else {
                        foreach ($searchString as $num => $cadena) {
                            $valor = "%" . $cadena . "%";
                            $valor2 = "" . $cadena . "%";

                            if ($objCampoF->TipoDato == 'numeric' || $objCampoF->TipoDato == 'decimal' || $objCampoF->TipoDato == 'integer' || strripos($objCampoF->TipoDato, 'int') !== false) {
                                //    echo $nomcamp.'--'.$valor;

                            } else {
                                $this->addCondicion($nomcamp, $operador, "'" . $valor . "'", $logicop, $reemplazo);
                            }
                            $reemplazo = '';
                        }
                    }
              }
          }
        }

    }


}

//3870
