<?php

/**
 * new XML READER
 * Object otoolriented version of old leoxml.php
 * Luis M. Melgratti
 */
class Histrix_XmlReader {

    /**
     * Constructor
     * @param string $dirXML
     * @param string $xml
     * @param bool $sub
     * @param string $xmlOrig
     * @param bool $subdir
     * @param bool objEmbebido
     */
    var $parameters;
    var $FilterContainer;
    var $Filter;

    function __construct($dirXML='', $xml='', $sub = false, $xmlOrig='', $subdir=null, $objEmbebido = false) {

        $this->dirXML = str_replace('//', '/',$dirXML );
        $this->xml = $xml;
        $this->sub = $sub;
        $this->fixXml = false;
        $this->xmlOrig = $xmlOrig;
        $this->subdir = $subdir;
        $this->objEmbebido = $objEmbebido;
        $registry = & Registry::getInstance();
        $this->i18n = $registry->get('i18n');
        $this->serialize = true;


        $this->Validator = new Histrix_TagValidator();

//        $this->Validator->readTagFile();


    }


    /**
     * add Referent Container to reader
     * @param ContDatos $dataContainer
     */
    public function addReferentContainer(&$dataContainer) {
        $this->ContenedorReferente = $dataContainer;
        $this->parentInstance = $dataContainer->parentInstance;
    }

    /**
     * Add Array or parameters to xmlReader
     * @param array $parameters
     */
    public function addParameters($parameters = '') {
        if ($parameters != '') {

            if (is_array($this->parameters))
                $this->parameters = array_merge($this->parameters, $parameters);
            else
                $this->parameters = $parameters;
        }
    }

    /**
     * Read and return Container
     */
    public function getContainer($serialized = false, $oldInstance = '') {

        $dataContainer = false;
        
        if ($oldInstance != '') {
            $dataContainer = $this->unserializeContainer( '', $oldInstance);
            return $dataContainer;
        }
        
        if ($dataContainer === false) {
            $dataContainer = $this->readXmlFile($this->dirXML, $this->xml, $this->sub, $this->xmlOrig, $this->subdir, $this->objEmbebido);
        }
        return $dataContainer;
    }

    /*
    // TODO improve garbage colector of this serialized containers
    public function getSerializedContainers() {
        return $this->serializados;
    }
    */
    /**
     *
     * @return ConDatos  Filter Container
     */
    public function getFilterContainer() {
        return $this->FilterContainer;
    }

    /**
     *
     * @return  field Filter
     */
    public function getFilterField() {
        return $this->Filter;
    }

    /**
     * get file from filesystem with fallbacks if not found
     * 
     * @param <string> $fileName field Name
     * @param <string> $mainDir  directory name
     * @param <string> $subdir   subdirectory
     * 
     * @return <string>
     */
    private function _getFile($fileName, $mainDir ='', $subdir='')
    {
        if ($subdir != '') {
            $dirXML2 = str_replace('//', '/', $mainDir . $subdir . '/');

            if (file_exists($dirXML2 . $fileName)) {
                $realPath = $dirXML2 . $fileName;
            } else {
                $realPath = $mainDir . $fileName;
            }
        } else {
            if (is_file($mainDir . $fileName))
                $realPath = $mainDir . $fileName;
        }
        return $realPath;
    }

    public function readXmlFile($dirXML, $xml, $sub = false, $xmlOrig='', $subdir=null, $objEmbebido = false) {
        $tipo = '';

        try {
            // get real Path for file
            $realPath = $this->_getFile($xml, $dirXML, $subdir);

           // prepare translator
           $this->translator = new Histrix_Translator($dirXML, $realPath);

            $this->realPath = $realPath;

            if (is_file($realPath)) {
                $histrixTag = simplexml_load_file($realPath);
            }

            if ($histrixTag == false) {
                throw new Exception($this->i18n['errorLoadXml'] . ' ' . $xml);
            }

            if (!($sub))
                $tipo = (string) $histrixTag['tipo'];

            // dashboard reading
            if ($tipo == 'dashboard') {
                $dashboard = $this->readDashboard($histrixTag);
                return $dashboard;
            }

            //$noform       = (string) $histrixTag['tipo'];
            $sololectura = (string) $histrixTag['sololectura'];

	    if ($_SESSION['readonly'] == 1){
		$sololectura = 'true';
	    }
            //$inserta = (string) $histrixTag['inserta'];

            
            $obsCont   = $this->translator->translate($histrixTag->obs);
            $customScript = (string) $histrixTag->customScript;

            foreach ($histrixTag->referentes as $ref) {
                $referentes[] = (string) $ref;
            }


            // get the Form Tag
            $MasterFormTag = isset($histrixTag->consulta) ? $histrixTag->consulta : $histrixTag->form;
            if (isset($histrixTag->consulta))
                $this->fixXml = true;
            foreach ($MasterFormTag as $formTag) {
                if (isset($formTag->tempTable)) {
                    foreach ($formTag->tempTable as $tempTable) {
                        $tableId = (string) $tempTable['id'];
                        $localTempTable[$tableId] = new TempTable($tableId);

                        // add relationships
                        foreach ($tempTable->field as $tempField) {
                            $localTempTable[$tableId]->addRelationship((string) $tempField['id'], (string) $tempField);
                        }
                        
                        // Add Rows
                        $rowNumber = 0;
                        foreach ($tempTable->row as $row) {

                            foreach ($row->field as $tempField) {
                                $localTempTable[$tableId]->addRow($rowNumber, (string) $tempField['id'], (string) $tempField);
                            }
                            $rowNumber++;
                        }
                    }
                }

                // Get Table Tag
                $tableTagArray = ($formTag->table) ? $formTag->table : $formTag->tabla;

                foreach ($tableTagArray as $tableTag) {

                    // Create Main Container

                    $dataContainer = new ContDatos((string) $tableTag['id'], (string) $tableTag['label'], (string) $histrixTag['tipo']);
                    $dataContainer->xml = $xml;
                    $dataContainer->idxml = str_replace('.', '_', $xml);

                    // Set External Temptable if available
                    if (isset($localTempTable))
                        $dataContainer->tempTables = $localTempTable;

                    // set Menu Identifier form call
                    if ($this->parameters['_menuId'] != '') {
                        $dataContainer->_menuId = $this->parameters['_menuId'];
                        $dataContainer->log = "true";
                    }

                    // set footer for pdf Printing
                    foreach ($histrixTag->pie as $pie) {
                        $dataContainer->pie['text'] = (string) $pie;
                        $dataContainer->pie['size'] = (string) $pie['size'];
                        $dataContainer->pie['align'] = (string) $pie['align'];
                        $dataContainer->pie['style'] = (string) $pie['style'];
                        $dataContainer->pie['color'] = (string) $pie['color'];
                    }

                    // Set Alias for Table
                    if (isset($tableTag['as']))
                        $dataContainer->alias = (string) $tableTag['as'];

                    // set xml Origin
                    if ($xmlOrig != '')
                        $dataContainer->xmlOrig = $xmlOrig;

                    if (isset($referentes))
                        $dataContainer->referentesAnexos = $referentes;

                    // show number of rows
                    if (isset($tableTag['showCantidad']))
                        $dataContainer->showCantidad = (string) $tableTag['showCantidad'];


                    // Conditional Saving
                    // name of the field that has de contitional value for proecssing xml
                    if (isset($tableTag['idCampoCond']))
                        $dataContainer->CampoCond = (string) $tableTag['idCampoCond'];

                    if (isset($tableTag['CondExp']))
                        $dataContainer->CampoCondExp = (string) $tableTag['CondExp'];

                    if (isset($tableTag['idCampoCant']))
                        $dataContainer->CampoCant = (string) $tableTag['idCampoCant'];

                    if (isset($tableTag['campoValorImputado']))
                        $campoValorImputado = (string) $tableTag['campoValorImputado'];

                    if (isset($tableTag['fieldValorImputado']))
                        $fieldValorImputado = (string) $tableTag['fieldValorImputado'];

                    if (isset($fieldValorImputado))
                        $campoValorImputado = $fieldValorImputado;
                    if (isset($campoValorImputado))
                        $dataContainer->campoValorImputado = $campoValorImputado;


                    // Accounting parameters
                    if (isset($tableTag['ajusteDebe']))
                        $dataContainer->ajusteDebe = (string) $tableTag['ajusteDebe'];
                    if (isset($tableTag['ajusteHaber']))
                        $dataContainer->ajusteHaber = (string) $tableTag['ajusteHaber'];


                    // Style for table and body
                    if (isset($tableTag['style']))
                        $dataContainer->styleTable = (string) $tableTag['style'];
//                    if (isset($tableTag['styleTbody']))
//                        $dataContainer->styleTbody = (string) $tableTag['styleTbody'];

                    // xml trigger Event
                    if (isset($histrixTag['evento']))
                        $dataContainer->evento = (string) $histrixTag['evento'];
                    // Eventos actions  XML
                    if ($histrixTag->eventos) {
                        foreach ($histrixTag->eventos as $eventos) {
                            foreach ($eventos->event as $event) {
                                $dataContainer->eventosXML[(string) $event['id']][(string) $event] = (string) $event;
                            }
                        }
                    }


                    // Confirmation Message
                    if (isset($histrixTag->confirmacion))
                        $dataContainer->confirmacion = (string) $histrixTag->confirmacion;

                    // confirmation button leyend
                    if ((string) $histrixTag['confirma'] != '')
                        $dataContainer->btnconfirma = $this->translator->translate($histrixTag['confirma']);

                    /**
                     *  Obtengo los parametros del tag <histrix> y los asigno al objeto Contenedor
                     */
                    foreach ($histrixTag->attributes() as $nompar => $valpar) {
                        $dataContainer->{$nompar} = (string) $valpar;
                        $this->Validator->fixAttributes('histrix', $nompar);
                    }

                    if ((string) $histrixTag['maxshowresult'] != '')
                        $dataContainer->maxshowresult = (string) $histrixTag['maxshowresult'];

                    // force Class
                    if (isset($histrixTag['class']))
                        $dataContainer->clase = (string) $histrixTag['class'];


                    // text results for batch processing
                    foreach ($histrixTag->txtresultados as $txtresultados) {
                        $dataContainer->prefijoResultados = (string) $txtresultados->prefijo;
                        $dataContainer->sufijoResultados = (string) $txtresultados->sufijo;
                    }


                    if ((string) $histrixTag['inserta'] != '')
                        $dataContainer->insertaABM = (string) $histrixTag['inserta'];

                    if ((string) $histrixTag['modifica'] != '')
                        $dataContainer->modificaABM = (string) $histrixTag['modifica'];

                        
                    if (isset($histrixTag['noform']))
                        $dataContainer->noForm = (string) $histrixTag['noform'];
                    if (isset($histrixTag['impresion']))
                        $dataContainer->xmlImpresion = (string) $histrixTag['impresion'];
                    if (isset($histrixTag['dirimpresion']))
                        $dataContainer->dirImpresion = (string) $histrixTag['dirimpresion'];
                    if (isset($histrixTag['graba']))
                        $dataContainer->grabaCab = (string) $histrixTag['graba'];
                    if (isset($histrixTag['grabafinal']))
                        $dataContainer->grabaFinal = (string) $histrixTag['grabafinal'];


                    if (isset($histrixTag['llenoreferente']))
                        $dataContainer->llenoReferente = (string) $histrixTag['llenoreferente'];

                    $dataContainer->tituloAbm = $this->translator->translate( isset($histrixTag->title) ? $histrixTag->title : $histrixTag->titulo); // replace old tag;;

                    if ($obsCont != '')
                        $dataContainer->obs = $obsCont;

                    if ($customScript != '') {
                        $dataContainer->customScript = str_replace('[__instance__]', $dataContainer->getInstance(), $customScript);
			
		    }


                    if (isset($this->parameters['xmlpadre']))
                        $dataContainer->xmlpadre = $this->parameters['xmlpadre'];

                    if (isset($this->parameters['dir']))
                        $dataContainer->dirxml = $this->parameters["dir"];
                    $dataContainer->dirXmlPrincipal = $dirXML;

                    if (!isset($dataContainer->xmlpadre))
                        $dataContainer->xmlpadre = $xml;
                    if ($sololectura != '')
                        $dataContainer->sololectura = $sololectura;


                    if (isset($this->parameters['_xmlreferente']))
                        $dataContainer->xmlReferente = $this->parameters['_xmlreferente'];

                    // referencia al xml externo que contiene el detalle
                    if (isset($formTag['formTbody']))
                        $dataContainer->formTbody = (string) $formTag['formTbody'];


                    foreach ($formTag->pdf as $pdfTag) {
                        if (isset($pdfTag->code)) {
                            $dataContainer->pdfCode = (string) $pdfTag->code;
                        }
                       
                        if (isset($pdfTag->header)) {
                            $dataContainer->pdfHeader = (string) $pdfTag->header;
                        }

                        if (isset($pdfTag->footer)) {
                            $dataContainer->pdfFooter = (string) $pdfTag->footer;
                        }
                    }

                    // referencia al xml externo que contiene el detalle
                    if (isset($formTag['detalle'])) {
                        $dataContainer->detalle = (string) $formTag['detalle'];
                        $dataContainer->iddetalle = str_replace('.', '_', $dataContainer->detalle);

                        $dataContainer->inline = (string) $formTag['inline'];
                        $dataContainer->inlineSingle = (string) $formTag['single'];
                        $dataContainer->hasDetail = (string) $formTag['hasDetail'];
                    }

                    // indica si muestro la cabecera del detalle a consultar
                    if (isset($formTag['showCab']))
                        $dataContainer->showAbm = (string) $formTag['showCab'];

                    // leo el xml y cargo el codigo a ejecutar en los inserts
                    foreach ($tableTag->ejecuta as $ejecuta) {
                        $dataContainer->codigoInsert = (string) $ejecuta;
                    }


                    // code to execute to fill table data
                    foreach ($tableTag->selectCode as $selectCode) {
                        $dataContainer->selectCode = (string) $selectCode;
                    }


                    // Custom on Row click javascript functions
                    foreach ($tableTag->onRowClick as $onRowClick) {
                        $dataContainer->onRowClick = (string) $onRowClick;
                    }

                    // Custom VALIDATIONS
                    $this->readValidations($tableTag, $dataContainer);

                    // Cargo las relaciones para las posibles Imputaciones
                    // TODO : DEPRECATED SOON
                    /*
                    foreach ($tableTag->imputacion as $imputacion) {

                        //$dataContainerImputacion =  leoXML($dirXML, (string) $imputacion['id'], true, $dataContainer->xml , $subdir);
                        $currentdir = $dirXML;
                        if (isset($imputacion['dir']))
                            $currentdir = (string) $imputacion['dir'];

                        $xmlReader = new Histrix_XmlReader($currentdir, (string) $imputacion['id'], true, $dataContainer->xml, $subdir);
                        $xmlReader->addParameters($this->parameters);
                        // falta contenedor referente?

                        $dataContainerImputacion = $xmlReader->getContainer();

                        $dataContainerImputacion->CampoValorImputado = (string) $imputacion->valor;
                        //$dataContainerImputacion->ColumnaImputar 		= (string) $imputacion->columna;
                        $dataContainer->Imputaciones[] = $dataContainerImputacion;
                    }
                        */
                    // leo el nombre del contenedor que va a ser llamado para cerrar el proceso

                    $this->closeProcess($tableTag, $dataContainer);

                    // Load Custom Selects
                    if (isset($tableTag->select))
                        $dataContainer->customSelect = (string) $tableTag->select;
                    if (isset($tableTag->sql))
                        $dataContainer->customSql = (string) $tableTag->sql;


                    // leo el xml y cargo el contenedor que contiene las tablas con movimientos a llenar
                    foreach ($tableTag->movimientos as $movimientos) {
                        // cargo la cabecera de movimientos
                        if (isset($movimientos['cabecera'])) {

                            $currentdir = $dirXML;
                            if (isset($movimientos['dir']))
                                $currentdir = (string) $movimientos['dir'];

                            $xmlReader = new Histrix_XmlReader($currentdir, (string) $movimientos['cabecera'], true, $dataContainer->xml, $subdir);
                            $xmlReader->addParameters($this->parameters);
                            $dataContainerCabecera = $xmlReader->getContainer();
                            $dataContainerCabecera->mainInstance = $dataContainer->getInstance();
                            $dataContainer->addCabecera($dataContainerCabecera);

                            /*
                            if (!isset($this->serializados))
                                $this->serializados[] = null;
                            // array_push( $this->serializados , $xmlReader->getSerializedContainers());
                            array_splice($this->serializados, count($this->serializados), 0, $xmlReader->getSerializedContainers());
                            */
                           
                           

                        }

                        foreach ($movimientos->relacion as $relacion) {

                            $currentdir = $dirXML;
                            if (isset($relacion['dir']))
                                $currentdir = (string) $relacion['dir'];


                            $xmlReader = new Histrix_XmlReader($dirXML, (string) $relacion['id'], true, $dataContainer->xml, $subdir, true);
                            $xmlReader->addParameters($this->parameters);
                            $xmlReader->addReferentContainer($this->ContenedorReferente);

                            // falta contenedor referente?
                            //
                            //loger('get cont'. (string) $relacion['id'], 'sql.log');
                            $dataContainerMovs = $xmlReader->getContainer();
                            //loger('fin get '. (string) $relacion['id'], 'sql.log');
                            //
                            // mark this as and sql generator xml
                            $dataContainerMovs->isSqlXml = true;
                            /*
                            $this->serializados[$dataContainer->xml . '_' . $dataContainerMovs->xml] = $dataContainer->xml . '_' . $dataContainerMovs->xml;
                            */
                            // remove (imputaciones)
                            $dataContainer->campoColumnaImputada = (string) $relacion->campoColumnaImputada;
                            if ((string) $relacion->fieldColumnaImputada != '')
                                $dataContainer->campoColumnaImputada = (string) $relacion->fieldColumnaImputada;

                            $dataContainer->campoValorImputado = (string) $relacion->campoValorImputado;
                            if ((string) $relacion->fieldValorImputado != '')
                                $dataContainer->campoValorImputado = (string) $relacion->fieldValorImputado;

                            $dataContainer->addCuerpoMov($dataContainerMovs);
                        }
                    }

                    // leo el xml y cargo el/los contenedor/es que contienen las preconsultas
                    foreach ($tableTag->preconsultas as $preconsultas) {
                        foreach ($preconsultas->xmlfile as $xmlfile) {
                            //$dataContainerPreCons =  leoXML($dirXML, (string) $xmlfile['id'], true, null, $subdir);
                            $xmlReader = new Histrix_XmlReader($dirXML, (string) $xmlfile['id'], true, null, $subdir);
                            $xmlReader->addParameters($this->parameters);
                            $dataContainer->addPreConsulta($xmlReader->getContainer());
                        }
                    }

                    $title = ($tableTag->title) ? $tableTag->title : $tableTag->titulo; // replace old tag
                    $titleString = (string) $title;
                    if ($titleString != '')
                        $dataContainer->setTitulo($titleString);

                    //FILTROS PREVIOS
                    foreach ($tableTag->filtro_ex as $filterEx) {
                        // Etiqueta del filtro
                        $label_filtro = $this->translator->translate((string) $filterEx->label);

                        foreach ($filterEx->tabla_ex as $tableEx) {

                            if ((string) $filterEx->tabla_ex["xml"] != '') {
                                $xmlReader = new Histrix_XmlReader($dirXML, (string) $filterEx->tabla_ex["xml"], true, null, $subdir);
                                $xmlReader->serialize = false;
                                $xmlReader->addParameters($this->parameters);
                                $this->FilterContainer = $xmlReader->getContainer();


                                $this->FilterContainer->xml = (string) $filterEx->tabla_ex["xml"];
                            }
                            else
                                $this->FilterContainer = new ContDatos((string) $tableEx["id"], '', 'consulta');

                            $this->FilterContainer->setTitulo($label_filtro);
                            $this->FilterContainer->CampoIcono = (string) $filterEx->tabla_ex["icono"];

                            // campos del filtro
                            $this->Filter = (string) $tableEx->filtra;

                            $fields = ($tableEx->field) ? $tableEx->field : $tableEx->campo; // replace old tag
                            // Load fields
                            foreach ($fields as $campo) {
                                $this->loadContainer($this->FilterContainer, $campo, $this->FilterContainer->TablaBase, $dirXML, $dataContainer, $subdir);
                                
                                //if ((string) $campo['muestra'] != '')
                                //    $muestraClave = (string) $campo['muestra'];
                            }

                            /* el grupo */
                            foreach ($tableEx->group as $group) {
                                $groupFields = ($group->field) ? $group->field : $group->campo; //replace old tag
                                foreach ($groupFields as $campoGrupo) {
                                    $tablagr = '';
                                    $campogr = (string) $campoGrupo['id'];

                                    if ($campoGrupo->expresion){
                                        $this->fixXml =true;
                                        $expressionGr = $campoGrupo->expresion;
                                    }
                                    else $expressionGr = $campoGrupo->expression;

                                    $expresion = (string) $campoGrupo->expresion;
                                    if ($expresion != '')
                                        $campogr = $expresion;

                                    if ((string) $campoGrupo['tabla'] != '')
                                        $tablagr = (string) $campoGrupo['tabla'] . '.';
                                    $this->FilterContainer->group[] = $tablagr . $campogr;
                                }
                            }

                            $newOrderTag = $tableEx->order;
                            if ($tableEx->orden != '') {
                                $this->fixXml = true;
                                $newOrderTag = $tableEx->orden;
                            }

                            foreach ($newOrderTag as $order) {
                                $orderField = ($order->field) ? $order->field : $order->campo; // replace old tag
                                foreach ($orderField as $campoOrden) {
                                    
                                    $orderType = isset($campoOrden['orderType'])?$campoOrden['orderType']:$campoOrden['tipoOrden'];
                                    $this->FilterContainer->setOrden((string) $campoOrden['id'], (string) $campoOrden['tabla'], (string) $orderType);
                                }
                            }
                        }
                    }

                    // Obtengo el contenedor Padre Referente
                    $dataContainerPadreRef = null;
                    if (isset($this->ContenedorReferente)) {
                        if ($this->ContenedorReferente->xmlpadre != '' && isset($this->parameters['_param_in'])) {
                            $dataContainerPadreRef = new ContDatos("");
                            $dataContainerPadreRef = Histrix_XmlReader::unserializeContainer(null, $this->ContenedorReferente->parentInstance);
                        }
                    }

		    // IMPORTANT RESETS IMPORT ARRAY
                    unset($dataContainer->importadatos);

                    // read field groups
                    //

                    if (isset($tableTag->fieldGroup))
                        foreach($tableTag->fieldGroup as $fieldGroup) {
//                        $fieldGroup = $tableTag->fieldGroup;

                            // Adding support for loops
                            $loop = 1;
                            $doLoop = false;

                            if (isset($fieldGroup->attributes()->repeat)){

                                $repeat = (string) $fieldGroup->attributes()->repeat;

                                // try to get field value
                                 
                                if ($dataContainer->getCampo($repeat)){
                                    $loop = $dataContainer->getCampo($repeat);
                                } else {
                                    $loop = $repeat;
                                }
                                
                                //$groupParameters['__repeat__'] = $loop;
                                $doLoop = true;
                               

                            }

                            for ($i=1; $i <= $loop; $i++) {

                                
                                // Read fieldGroup parameters to pass to each Field
                                unset($groupParameters);

                                if ($doLoop){
                                    //echo $i;
                                    $groupParameters['__repeat__'] = $i;
                                }

                                foreach ($fieldGroup->attributes() as $nompar => $valpar) {
                                    if ($nompar != 'id') {
                                        $groupParameters[$nompar] = (string) $valpar;
                                    }
                                }
                                $processGroup = false;
                                // test field Group condition
                                if (isset($fieldGroup->if)) {
                                    $if = (string) $fieldGroup->if;
                                    $processGroup = eval($if);
                                } else {
                                    // if there is no condition the fieldGroup will be processed anyway
                                    $processGroup = true;
                                }

                                if ($processGroup) {
                                    $this->readFields($fieldGroup , $dirXML, $subdir, $dataContainer , $dataContainerPadreRef, $groupParameters ) ;
                                    $this->readValidations($fieldGroup, $dataContainer);
                                    $this->closeProcess($fieldGroup, $dataContainer);
                                }

                            }


                        }



                    // Read Fields
                    $this->readFields($tableTag , $dirXML, $subdir, $dataContainer , $dataContainerPadreRef) ;


                    /* Filtros NO obligatorios Preestablecidos */
                    /* if ($tabla->filtros != '') {
                      $this->Validator->fixTags($dirXML.$subdir.'/'.$dataContainer->xml);
                      } */
                    $filters = ($tableTag->filtros) ? $tableTag->filtros : $tableTag->filters;
                    if (isset($filters)){
                    
                	$dataContainer->autoFilter = (string) $filters['auto'];
                	
                        foreach ($filters as $filtros) {
                            $filterFields = ($filtros->field) ? $filtros->field : $filtros->campo;
                            foreach ($filterFields as $campofiltro) {

                                $valor = (string) $campofiltro->valor; /* casos especiales */
                                $valor = Types::checkToday($valor, 'date');

                                if ((string) $campofiltro->valor['eval'] == 'true') {
                                    $val = '$valor = ' . $valor . ' ;';
                                    eval($val);
                                }

                                foreach ($campofiltro->parametro as $par) {
                                    //$dataContainerParametro = leoXML($dirXML, (string) $par['id'], true, null, $subdir);
                                    $parxml    = ($par['xml'] != '')?   (string) $par['xml']  : (string) $par['id'];
                                    $tablename = ($par['table'] != '')? (string) $par['table']: '' ;

                                    if ($parxml != ''){

                                        $xmlReader = new Histrix_XmlReader($dirXML,  $parxml, true, $dataContainer->xml, $subdir);
                                        $xmlReader->addParameters($this->parameters);
                                        $dataContainerParametro = $xmlReader->getContainer();

                                    } else {

                                        $dataContainerParametro = new ContDatos($tablename, 'param', 'consulta');
                                    }

                                    $dataContainerParametro->retorna = (string) $par['retorna'];

                                    $paramterFields = ($par->field) ? $par->field : $par->campo; //replace old tag
                                    foreach ($paramterFields as $campopar) {
                                        $this->loadContainer($dataContainerParametro, $campopar, '', $dirXML, $dataContainer, $subdir);
                                    }
                                    $dataContainerParametro->cargoCampos();
                                    $valor = $dataContainerParametro->getCampo($dataContainerParametro->retorna)->getValor();
                                }

                                $modificador = (string) $campofiltro->operador['modificador'];
                                $oplogico = (string) $campofiltro->operador['oplogico'];
                                if ($oplogico == '') {
                                    $oplogico = 'and';
                                }
                                $dataContainer->getCampo((string) $campofiltro['id'])->oplogico = $oplogico;

                                // TODO
                                // REPLACE THIS with proper field data
                                $filterObject = new filtro((string) $campofiltro['id'],
                                        (string) $campofiltro->operador,
                                        $this->translator->translate((string) $campofiltro->label),
                                        $valor, 'xml', $modificador,
                                        (string) $campofiltro['grupo'],
                                        (string) $campofiltro['deshabilitado'],
                                        (string) $campofiltro['copia']);
                                if ($campofiltro)
                                    foreach ($campofiltro->attributes() as $nompar => $valpar) {
                                        $filterObject->{$nompar} = (string) $valpar;
                                        
                                        $this->Validator->fixAttributes('field', $nompar);
                                    }

                                // Label and style parameters
                                if ($campofiltro->label)
                                    foreach ($campofiltro->label->attributes() as $nompar => $valpar) {
                                        $filterObject->{$nompar} = (string) $valpar;
                                        $this->Validator->fixAttributes('label', $nompar);
                                    }
                                $dataContainer->filtros[] = $filterObject;


                                // Select default Data for Filters
                                $refereceObject = $dataContainer->getCampo((string) $campofiltro['id']);
                                if ($valor == '') {
                                    // if the Object has an input select box
                                    if (is_array($refereceObject->opcion)) {
                                        $valor = current(array_keys($refereceObject->opcion));
                                    }
                                }
                                // Add Conditions
                                if ($valor != '') {
                                    $dataType = Types::getTypeXSD($refereceObject->TipoDato, 'xsd:integer');
                                    if ($dataType == 'xsd:integer' || $dataType == 'xsd:decimal') {

                                        $dataContainer->addCondicion((string) $campofiltro['id'], (string) $campofiltro->operador, $valor, ' ' . $oplogico . ' ', '', 'false');
                                    } else {
                                        $dataContainer->addCondicion((string) $campofiltro['id'], (string) $campofiltro->operador, $valor, ' ' . $oplogico . ' ', '');
                                    }

                                    $dataContainer->setCampo((string) $campofiltro['id'], $valor);
                                    $dataContainer->setNuevoValorCampo((string) $campofiltro['id'], $valor);
                                }
                            }
                        }
		    } // end filters
		    
                    /* el grupo */
                    foreach ($tableTag->group as $group) {
                        $groupFields = ($group->field) ? $group->field : $group->campo; // replace old tag
                        foreach ($groupFields as $campoGrupo) {
                            $tablagr = '';
                            $campogr = (string) $campoGrupo['id'];

                            if ($campoGrupo->expresion){
                                $this->fixXml =true;
                                $expressionGr = $campoGrupo->expresion;
                            }
                            else $expressionGr = $campoGrupo->expression;
                            
                            $expresion = (string) $expressionGr;
                            if ($expresion != '')
                                $campogr = $expresion;

                            if ((string) $campoGrupo['tabla'] != '')
                                $tablagr = (string) $campoGrupo['tabla'] . '.';
                            $dataContainer->group[] = $tablagr . $campogr;
                        }
                    }

                    /* el orden */

                    $newOrderTag = $tableTag->order;
                    if ($tableTag->orden != '') {
                        $this->fixXml = true;
                        $newOrderTag = $tableTag->orden;
                    }

                    foreach ($newOrderTag as $orden) {
                        $orderFields = ($orden->field) ? $orden->field : $orden->campo; // replace old tag
                        foreach ($orderFields as $campoOrden) {
                            $idCampo = (string) $campoOrden['id'];

                            $orderType = isset($campoOrden['orderType'])?$campoOrden['orderType']:$campoOrden['tipoOrden'];                                                

                            if ($idCampo != ''){
                                $dataContainer->addOrden($idCampo, (string) $campoOrden['tabla'], (string) $orden['ordenaTemporal'], (string) $orderType);
                            }
                        }
                    }


                    // Creo los Gráficos
                    foreach ($tableTag->grafico as $grafico) {
                        $idxml = str_replace('.', '_', $xml);

                        $inst = $dataContainer->getInstance();
                        $idGraf = $idxml .  str_replace(' ', '_', (string) $grafico['id']); 



	                $graf['ancho'] = (string) $grafico->ancho;
                        $graf['alto'] = (string) $grafico->alto;
                        $graf['tipo'] = (string) $grafico->tipo;
                        if ((string) $grafico->datos != '')
                            $graf['datos'] = (string) $grafico->datos;

                        if ((string) $grafico->etiquetas != '')
                            $graf['etiquetas'] = (string) $grafico->etiquetas;

                        $title = ($grafico->title) ? $grafico->title : $grafico->titulo; // replace old tag

                        $graf['titulo'] = (string) $title;
                        $graf['subtitulo'] = (string) $grafico->subtitulo;


                        $graf['objective'] = (string) $grafico->objective;
                        $graf['option_name'] = 'CHART::'.$idGraf;


                        $htxParameter = new Histrix_Parameter($graf['option_name'], 'Objetivo grafico '.$title, $graf['objective']);
                        $graf['objective'] = $htxParameter->getValue();


                        foreach ($grafico->serie as $serie) {
                            $graf['series'][(string) $serie] = (string) $serie;
                        }

                        $dataContainer->grafico[$idGraf] = $graf;
                        unset($graf);
                    }
                } // END table


                /* JOINS */
                foreach ($formTag->join as $join) {

                    $joinTable = ($join->table) ? $join->table : $join->tabla;  // replace old tag
                    foreach ($joinTable as $tableTag) {

                        $joinContainer = new ContDatos(strtr((string) $tableTag['id'], "-", "_"), (string) $tableTag['label']);
                        $joinContainer->grupo = (string) $join['grupo'];
                        $joinContainer->alias = (string) $tableTag['as'];
                        $joinFields = ($tableTag->field) ? $tableTag->field : $tableTag->campo;  // replace old tag
                        foreach ($joinFields as $campo) {
                            $joinTblName = (isset($joinContainer->nombreTabla)) ? $joinContainer->nombreTabla : '';
                            $this->loadContainer($joinContainer, $campo, $joinTblName, $dirXML, null, $subdir);
                        }
                        /* Le agrego el subcontenedor */
                        $dataContainer->addJoin($joinContainer, (string) $join['tipo']);
                    }
                }

                $arrayContainers[] = $dataContainer;
            } // end form
        } catch (Exception $ex) {
            errorMsg($ex->getMessage());
            //$error = true;
        }

        // UNIONS
        if ($arrayContainers != '')
            foreach ($arrayContainers as $num => $cont) {
                if ($num > 0) {
                    $arrayContainers[0]->unionContainers[] = $cont;
                }
            }

        if ($tipo != 'ing') {
            $isSubxml = false;
            if (isset($this->parameters['xmlsub']))
                $isSubxml = ($this->parameters['xmlsub'] == 'true') ? true : $isSubxml;

            // si le saco esto no anda la impresion del pdf, ver por que
            if ($arrayContainers[0]->tipoAbm == 'ficha' &&
                    $arrayContainers[0]->sololectura == 'true' ||
                    $isSubxml) {

                if ($objEmbebido == false && is_object($arrayContainers[0]))
                    $arrayContainers[0]->Select();
            }
        }

        if ($this->serialize) {

            $serializedId = Histrix_XmlReader::serializeContainer($arrayContainers[0]);
            /*
            if ($serializedId != '')
                $this->serializados[$serializedId] = $serializedId;
            */
        }

        if ($this->fixXml) {
            $this->Validator->fixTags($dirXML . $subdir . '/' . $dataContainer->xml);
        }

        return $arrayContainers[0];
    }


    /**
     * Read Validations
     */
    private function readValidations($tableTag, &$dataContainer) {
        foreach ($tableTag->validation as $validation) {
            $optional = (string) $validation['optional'];
            if ($optional == '')
                $optional = 'false';
            $condition = (string) $validation->condition;

            $dataContainer->validationType[$condition] = $optional;
            $dataContainer->validations[$condition] = (string) $validation->message;
        }
    }

    /**
     * Close Process
     */
    private function closeProcess($tableTag, &$dataContainer) {
        foreach ($tableTag->cierraproceso as $cierraproceso) {
            $dataContainer->cierraproceso = (string) $cierraproceso['id'];
            $dataContainer->cierraprocesoDir = (string) $cierraproceso['dir'];
            $dataContainer->cierraprocesoCondition = (string) $cierraproceso['condition'];

            foreach ($cierraproceso->paring as $paring) {
                $strparingdest = (string) $paring['destino'];
                $strparing = (string) $paring;
                $paringdest = ($strparingdest != '') ? $strparingdest : $strparing;

                $dataContainer->paring[$paringdest] = (string) $paring;
            }
        }
    }

    /**
     *  Read fields form Table
     */
    private function readFields($tableTag , $dirXML, $subdir, $dataContainer,  $dataContainerPadreRef, $groupParameters=null) {
        // RECORRO LOS CAMPOS
        $fields = ($tableTag->field) ? $tableTag->field : $tableTag->campo;
        foreach ($fields as $campo) {

            //TODO PASAR GROUP PARAMETERS
            // Fetch field from another query resultset
            if (isset($campo->fields)) {
                foreach ($campo->fields as $innerfields) {

                    $xmlReader = new Histrix_XmlReader($dirXML, (string) $innerfields["xml"], true, $dataContainer->xml, $subdir);
                    $xmlReader->addParameters($this->parameters);
                    $fieldContainer = $xmlReader->getContainer();

                    $fieldRecordSet = $fieldContainer->Select();

                    /* recorro los campos */
                    $i = 0;
                    if ($fieldRecordSet) {

                        while ($row = _fetch_array($fieldRecordSet)) {
                            foreach ($row as $fieldName => $fieldValue) {
                                $fieldObj = $fieldContainer->getCampo($fieldName);

                                if ($fieldObj->oculto == 'true')
                                    continue;
                                // field creation

                                $field = $campo;

                                $field['id'] = $i . $fieldValue;
                                $field->label = $fieldValue;
                                $field->help = $fieldValue;
                                $field->parentField = &$campo;

                                $this->loadContainer($dataContainer, $field, $dataContainer->TablaBase, $dirXML, $dataContainerPadreRef, $subdir, $groupParameters);

                            }
                            $i++;
                        }
                    }
                }
            }
            else {

                // Load Container with fields
                $this->loadContainer($dataContainer, $campo, $dataContainer->TablaBase, $dirXML, $dataContainerPadreRef, $subdir, $groupParameters);
            }
        }
        
        
                            // lleno un array con los datos del xml que se usa para importar datos al ingreso
                    foreach ($tableTag->importar as $importar) {
                        $importField = ($importar->field) ? $importar->field : $importar->campo; //replace old tag
                        foreach ($importField as $campo) {
                            $valorParam = (string) $campo;
                            if ($valorParam == '')
                                $valorParam = (string) $campo['id'];
                            $parametro[(string) $campo['id']] = (string) $campo;
                        }
                        $xmlImportar = (string) $importar['xml'];
                        
                        $dataContainer->importadatos[$xmlImportar]['campos'] = $parametro;
                        $dataContainer->importadatos[$xmlImportar]['accesskey'] = (string) $importar['accesskey'];
                        $dataContainer->importadatos[$xmlImportar]['label'] = utf8_decode((string) $importar['label']);
                        $dataContainer->importadatos[$xmlImportar]['titulo'] = (string) $importar['titulo'];
                        $dataContainer->importadatos[$xmlImportar]['boton'] = (string) $importar['boton'];
                        $dataContainer->importadatos[$xmlImportar]['dir'] = (string) $importar['dir'];
                        $dataContainer->importadatos[$xmlImportar]['width'] = (string) $importar['width'];
                        //$this->serializados[$dataContainer->xml . '_' . $xmlImportar] = $dataContainer->xml . '_' . $xmlImportar;
                        unset($parametro);
                    }


        
    }

    /**
     * Get Serialized Id from Container
     */
    public static function getSerializedId(&$dataContainer, $xmlContext='', $instance='') {

        return $dataContainer->getInstance();
        /*if (is_object($dataContainer)) {
            if ($instance == '')
                $instance = $dataContainer->getInstance();

            $xmlName = $dataContainer->xml;
            $xmlContext = isset($dataContainer->xmlContext) ? $dataContainer->xmlContext : $xmlContext;

            if ($xmlContext == '') {
                $xmlContext = isset($dataContainer->xmlOrig)  ? $dataContainer->xmlOrig  : $xmlContext;
                $xmlContext = isset($dataContainer->xmlpadre) ? $dataContainer->xmlpadre : $xmlContext;
            }
        } else {
            $xmlName = $dataContainer;
        }

        // instance ID method
        if ($instance != '') {
            if ( isset($_SESSION['instances'][$instance])) {
                return $_SESSION['instances'][$instance];
            }
        }


        if ($xmlContext == $xmlName)
            $xmlContext = '';

        // $xmlContext = '';
        $serializedId = $xmlContext . '_' . $xmlName;
        return $serializedId;
        */
    }

    /**
     * serialize Container with unique ID to improve garbage collector
     * @param ContDatos $dataContainer
     * @return string
     */
    public static function serializeContainer(&$dataContainer) {


        $serializedId = $dataContainer->getInstance();

        if (function_exists('igbinary_serialize')){
            /// IGBINARY
            $_SESSION['xml'][$serializedId] = igbinary_serialize($dataContainer);

        } else {
            // Normal
            $_SESSION['xml'][$serializedId] = serialize($dataContainer);
        }

        return $serializedId;
    }

    /**
     * unserialize Container with Unique ID to improve garbage collector
     * @param string / ContDatos $xml
     * @param string $xmlContext
     * @return ContDatos
     */
    public static function unserializeContainer($dataContainer=null, $serializedId ='') {
        if (is_object($dataContainer)){
            $serializedId = $dataContainer->getInstance();
        }

        

        if (isset($_SESSION['xml'][$serializedId])) {

            if (function_exists('igbinary_unserialize')){
                return igbinary_unserialize($_SESSION['xml'][$serializedId]);
            } else {
                return unserialize($_SESSION['xml'][$serializedId]);
            }


        } else {
            $result = false;
            loger('error deserializing:' . $serializedId . ' instance: ' . $instance, 'serialization.log');

            return $result;

        }
    }

    function loadContainer($dataContainer, $campo, $tabla = '', $dirXML='', $dataContainerPadre=null, $subdir = null, $groupParameters=null) {

	    ///////////// test field eval condition  ///////////////
        if (isset($campo->eval)) {
              $evalif = (string) $campo->eval;
              $processField = eval($evalif);
        } else {
              // if there is no condition the field will be processed anyway
              $processField = true;
        }

		if ($processField != true) return;
        /////////////////////////////////////////////////////////

        if ($subdir != null)
            $dataContainer->subdir = $subdir;


        // append order in field name
        $sufijo = '';
        if (isset($groupParameters['__repeat__'])){
            $sufijo = '_'.$groupParameters['__repeat__'];
        }
        $idCampo = (string) $campo['id'];
        $idCampo .= $sufijo;

        $idTabla = (string) $campo['tabla'];
        $local = (string) $campo['local'];

        $expresion = '';
        if ($campo->expresion){
            $this->fixXml =true;
            $expression = $campo->expresion;
        }
        else $expression = $campo->expression;

        if ($expression != '')
        $expresion = (string) $expression;

        $dataContainer->addCampo($idCampo, $expresion, '', '', $tabla, $local);

        $refField = $dataContainer->getCampo($idCampo);

        
        if(isset($groupParameters['__repeat__'])){
            $refField->__repeat__ = $groupParameters['__repeat__'];
        }
        

        // add default options from fieldGroup attributes
        if ($groupParameters != null) {
            foreach($groupParameters as $attrName => $attrValue) {
                if ($attrName != 'id') {
                    $refField->{$attrName} = $attrValue;
                    $this->Validator->fixAttributes('field', $attrName);
                    // Metodo Nuevo los pongo todos en el mismo array (NO SE CUAL DEJAR)
                    $refField->Parametro[strtolower($attrName)] =  $attrValue;
                }
            }
        }


        $refField->_DataContainerRef = &$dataContainer;

        if ($campo->expresion){
            $this->fixXml =true;
            $expression = $campo->expresion;
        }
        else $expression = $campo->expression;

        if ($expression != '') {
            $refField->having       = (string) $expression['having'];
            $refField->expresionSql = (string) $expression['sql'];
        }

        if ($idTabla != '')
            $refField->TablaPadre = $idTabla;
        if (isset($dataContainer->alias))
            $refField->alias = $dataContainer->alias;

        if (isset($campo->invalid))
            $refField->invalid = (string) $campo->invalid;

        /**
         *  	Obtengo los parametros del tag <campo> y los asigno al objeto Campo
         */
        foreach ($campo->attributes() as $nompar => $valpar) {
            $refField->{$nompar} = (string) $valpar;
            $this->Validator->fixAttributes('field', $nompar);

            if ($nompar == 'break')
                $dataContainer->hasBreak = 'true';

            // Metodo Nuevo los pongo todos en el mismo array (NO SE CUAL DEJAR)
            $refField->Parametro[strtolower($nompar)] = (string) $valpar;
        }

        if (isset($refField->ldifName))
            $dataContainer->ldifExport = true;

        if (isset($campo['id_temp']))
            $refField->campoTemp = (string) $campo['id_temp'];
        if (isset($campo['id_cab']))
            $refField->campoCab = (string) $campo['id_cab'];
        if (isset($campo['id_padre']))
            $refField->campoPadre = (string) $campo['id_padre'];
        if (isset($campo['id_int']))
            $refField->campoInterno = (string) $campo['id_int'];
        if (isset($campo['id_cont']))
            $refField->contenedorInterno = (string) $campo['id_cont'];



        if ((string) $campo->size != '')
            $dataContainer->setSize($idCampo, (string) $campo->size, '', '', $tabla);
        if ((string) $campo->maxsize != '')
            $refField->maxsize = (string) $campo->maxsize;


        $lbl = $campo->label;
  /*      $valLabel = (string) $campo->label;

        foreach($campo->label as $indexLbl => $langlabel){
            $lbl = $langlabel;
            
            if ($_SESSION['lang'] == (string) $lbl['lang'])
                $valLabel = (string) $lbl;
        }
*/

        $refField->Etiqueta = $this->translator->translate($campo->label).str_replace('_', ' ', $sufijo);


        if (isset($campo->htmllabel))
            $refField->Htmllabel = (string) $campo->htmllabel;

        $dataContainer->tablas[$dataContainer->TablaBase]->etiquetas_reverse[(string) $campo->label] = $idCampo;

        if (isset($campo->label['xml']))
            $refField->xmletiq = (string) $campo->label['xml'];

        if (isset($campo->preparam)) {
            $refField->preparam = (string) $campo->preparam;
            $refField->preparamOP = (string) $campo->preparam['operador'];

            $preparamCampo = (string) $campo->preparam['campo'];
            $preparamField = (string) $campo->preparam['field'];
            if ($preparamField != '')
                $preparamCampo = $preparamField;
            $refField->preparamCampo = $preparamCampo;
        }
        if (isset($campo->prefijo))
            $refField->prefijo = (string) $campo->prefijo;

        if (isset($refField->rss)) {
            $dataContainer->rss = true;
        }
        if (isset($refField->esClave))
            if ($refField->esClave == 'true') {
                //$refField->esClave = (string) $campo['esclave'];
                if ($dataContainer->tipoAbm == 'ing' || $dataContainer->tipoAbm == 'grid')
                    $dataContainer->TablaTemporal->indices[$idCampo] = $idCampo;
            }


        //LABELS

        /**
         *  Obtengo los parametros del tag <label> y los asigno
         *  al objeto Campo (deberia Asignarse a OTRO)
         */
        if ($lbl)
            foreach ($lbl->attributes() as $nompar => $valpar) {
                $refField->{$nompar} = (string) $valpar;
                $this->Validator->fixAttributes('label', $nompar);

            }
        foreach ($campo->habilita as $habilita) {
            $refField->habilita = (string) $habilita;
        }
        foreach ($campo->conditionalDisplay as $conditionalDisplay) {
            $refField->conditionalDisplay = (string) $conditionalDisplay;
        }

        /*         * **
         * Read Import TAG
         * ** */

        foreach ($campo->importar as $importar) {
            $impXml = (string) $importar['xml'];
            $impLabel = (string) $importar['label'];
            $importField = ($importar->field) ? $importar->field : $importar->campo; // replace old tag

            foreach ($importField as $importarCampo) {
                $impCampos[(string) $importarCampo['id']] = null;
            }
            $imp['xml'] = $impXml;
            $imp['dir'] = (string) $importar['dir'];
            $imp['label'] = $impLabel;
            $imp['campos'] = $impCampos;

            $refField->importacion = $imp;
//            $this->serializados[$dataContainer->xml . '_' . $impXml] = $dataContainer->xml . '_' . $impXml;
        }




        if ((string) $campo->tipo != '') {
            $dataContainer->setTipo($idCampo, (string) $campo->tipo, '', (string) $campo->tipo['aletras']);
            if (isset($campo->tipo['Type']))
                $refField->FType = (string) $campo->tipo['Type'];
            if (isset($campo->tipo['encode']))
                $refField->FEncode = (string) $campo->tipo['encode'];
            if (isset($campo->tipo['url']))
                $refField->url = (string) $campo->tipo['url'];
        }

        $refField->setOculto((string) $campo['oculto']);


        /* if (isset($campo->logReference)){
          $dataContainer->logReference = $idCampo;
          } */


        // set custom format
        if ($campo->formato)
            $this->fixXml = true;

        $format = ($campo->format) ? $campo->format : $campo->formato; // replace old tag          

        if ($format != '')
            $refField->Formato = (string) $format;


        if (isset($campo->help)){

            $refField->ayuda = $this->translator->translate($campo->help);
        }
        //$refField->setOculto((string) $campo['oculto']);
        //$dataContainer->setOculto($idCampo, (string) $campo['oculto']);
        //$dataContainer->setArbol($idCampo, (string) $campo['arbol']);
        if (isset($campo['arbol']))
            $refField->Arbol = (string) $campo['arbol'];



        foreach ($campo->jsextract as $extract) {
            $posiciones['posini'] = (string) $extract['posini'];
            $posiciones['posfin'] = (string) $extract['posfin'];
            $refField->jsextract[(string) $extract] = $posiciones;
        }

        // Custom Javascript Function
        foreach ($campo->jsfunction as $jsfunction) {
            $event = (string) $jsfunction['event'];
            $function = (string) $jsfunction;
            $refField->jsfunction[$event][] = $function;
        }

        //jquery calculation php implementation
        foreach ($campo->calculate as $calculate) {
            $refField->calculateStrings[] = (string) $calculate;
        }

        // Evaluacion Javascript
        foreach ($campo->jseval as $jseval) {
            $operatorsOld = array('(', ')', '[', ']', '?', ':', '+', '-', '*', '/', '==', '!=', "\\'", '  ');
            $operatorsNew = array('( ', ' )', '[ ', ' ]', ' ? ', ' : ', ' + ', ' - ', ' * ', ' / ', ' == ', ' != ', " \\' ", ' ');

            $campodestino = (string) $jseval['campodestino'];

            $fielddestino = (string) $jseval['fielddestino'];
            if ($fielddestino != '')
                $campodestino = $fielddestino;

            $tot = (string) $jseval['total'];
            if ($campodestino == '')
                $campodestino = '__EVAL';

            $formula = (string) $jseval;
            //echo $campodestino.'__'.$formula.'<br>';
            // sanitize formula on load
            $formula = str_replace($operatorsOld, $operatorsNew, $formula);

            //echo '|'.$formula.'<br>';


            if ($tot == 'true')
                $refField->jsevaltotal[$campodestino] = $formula; // para las grillas
            else
                $refField->jseval[$campodestino] = $formula; // para los abms
            $noact = (string) $jseval['actxml'];
            if ($noact != '')
                $refField->jsevalactxml[$campodestino] = $noact;
            if ($noact == 'false')
                $refField->jsevalactxml[$campodestino] = 'false';
        }

        // propagacion de valores o otros XMLs
        /* 	if ($campo->propagate != '')
          foreach ($campo->propagate as $propagate) {
          $destino = (string) $propagate['xml'];
          $refField->propagate[$destino] = (string) $propagate;

          } */

        // Evaluacion Javascript webforms 2.0
        foreach ($campo->onformchange as $onformchange) {
            $refField->onformchange = (string) $onformchange;
        }


        foreach ($campo->detalle as $detalle) {
            $refField->setDetalle((string) $detalle);
            //$dataContainer->setDetalle($idCampo, (string) $detalle);
        }
        // Guardo en elCampo de donde sacar los datos para pasarselos
        // como parametros de ingreso
        // al contenedor hijo que este campo contiene
        foreach ($campo->paring as $paring) {
            $destinationField = (string) $paring['destino'];
            $operador = (string) $paring['operador'];
            $reemplazo = (string) $paring['reemplazo'];
            $refField->paring[$destinationField]['valor'] = (string) $paring;
            $refField->paring[$destinationField]['operador'] = $operador;
            $refField->paring[$destinationField]['reemplazo'] = $reemplazo;
        }

        foreach ($campo->titulo as $titulo) {
            $refField->titulo = (string) $titulo;
        }

        foreach ($campo->pdfancho as $ancho) {
            $refField->pdfancho = (string) $ancho;
        }

        if ((string) $campo->label['PDFancho'] != '')
            $refField->pdfancho = (string) $campo->label['PDFancho'];

        if ((string) $campo->label['size'] != '')
            $refField->lblsize = (string) $campo->label['size'];

        if ((string) $campo->label['pdffill'] != '')
            $refField->pdffill = (string) $campo->label['pdffill'];

        if ((string) $campo->label['posx'] != '') {
            $refField->posX = (string) $campo->label['posx'];
        }
        if ((string) $campo->label['posy'] != '') {
            $refField->posY = (string) $campo->label['posy'];
        }


        $disabled = ((string) $campo['deshabilitado'] != '') ? (string) $campo['deshabilitado'] : (string) $campo['disabled'];

        if ($disabled != '')
            $refField->deshabilitado = $disabled;

        $dataContainer->addSuma($idCampo, (string) $campo['suma']);
        $dataContainer->addAcumula($idCampo, (string) $campo['acumula']);

        //   $refField->retroalimenta = (string) $campo['retroalimenta'];

        if (isset($this->parameters['_param_in'])) {
            $inParametersArray = $this->parameters['_param_in'];
            
            $hay = false;
            if (array_key_exists($idCampo, $inParametersArray)) {
                $inParameter = $inParametersArray[$idCampo];
                $hay = true;
            } else {
                if (in_array($idCampo, $inParametersArray)) {
                    $inParameter = $idCampo;
                    $hay = true;
                }
            }

            if ($hay) {

                // Busco primero en el Contenedor principal
                $encuentra = false;
                $campodelReferente = null;
                if (isset($this->ContenedorReferente))
                    $campodelReferente = $this->ContenedorReferente->getCampo($inParameter);
    
                if ($campodelReferente != null) {

//    loger($campodelReferente->NombreCampo.' == '.$campodelReferente->valor, 'param_in');   
                            
                    $valorReferente = $campodelReferente->valor;

                    $encuentra = true;
                    if ($campodelReferente->suma == 'true' && $campodelReferente->forceCellValue != 'true') {
                        $valorReferente = $campodelReferente->Suma;
                        
                    } else {
                        /* busco en la tabla temporal */
                        // Get value from inner table
                        if (is_array($this->ContenedorReferente->TablaTemporal->Tabla[$this->parameters['__row']])) {
                            
                            $dataFromRow = $this->ContenedorReferente->TablaTemporal->Tabla[$this->parameters['__row']];
                            $valorReferente = $dataFromRow[$inParameter];

                            if (isset($dataFromRow[$inParameter])){
                        	$encuentra = true;
                            }
//loger($this->ContenedorReferente->TablaTemporal->Tabla, 'param_in');
//loger('Tabla == '.$valorReferente, 'param_in');                               
                            
                        }
                    }
                    //  echo 'valor del campo referente 1'.$inParameter.'='.$valorReferente;
                }

                
//loger($inParameter.' = '.$valorReferente, 'param_in');                               

                // Busco en las Cabeceras del contenedor principal
                if ($this->ContenedorReferente->CabeceraMov)
                    foreach ($this->ContenedorReferente->CabeceraMov as $cabecera) {
                        $cabecera = Histrix_XmlReader::unserializeContainer($cabecera);
                        $headerField = $cabecera->getCampo($inParameter);
                        if ($headerField != null) {
                            $valorReferente = $headerField->valor;
                            $encuentra = true;
                        }
                    }

                     
                    
                // Busco en el Padre del Referente
                if ($encuentra == false) {
                    if ($this->ContenedorReferente->xmlpadre != '') {
                        if ($dataContainerPadre != '') {
                            $campoReferente = $dataContainerPadre->getCampo($inParameter);

                            if ($campoReferente != null) {
                                $valorReferente = $campoReferente->valor;
                                $encuentra = true;
                                //unset($campoReferente);
                            }
                        } else {
                    	// grid method for get parent container
                    	
//                    	echo  $this->ContenedorReferente->xml.'___'. $this->ContenedorReferente->parentInstance.'___';

                            /*                            
                            echo 'xml'.$this->xml;
                            echo '<br>xmlreferente '.$this->ContenedorReferente->xml;
                            echo '<br>xmlreferente->padre '.$this->ContenedorReferente->xmlpadre;
                            echo '<br>in parameter ' .$inParameter;

                            flush();
                            die();
                            */
                            ////////////////
                            // NOT WORKING
                            // 
                            $dataContainerPadreRef = new ContDatos("");
                            $dataContainerPadreRef = Histrix_XmlReader::unserializeContainer(null, $this->ContenedorReferente->parentInstance);
                            if ($dataContainerPadreRef != '') {

                        	   $campoReferente = $dataContainerPadreRef->getCampo($inParameter);
                        	
                        	   if ($campoReferente != null) {
                            	   $valorReferente = $campoReferente->valor;
                        	       $encuentra = true;
                            	   //unset($campoReferente);
                                }
                        	}

                        }
                    }
                }


                
                if (isset($valorReferente)) {
//                    echo 'valor del campo '.$idCampo.'  DEL referente: '.$inParameter.'='.$valorReferente;
                    $dataContainer->addCondicion($idCampo, '=', "'".$valorReferente."'", 'and', 'reemplazo', 'false');

                    /*$dataContainer->setCampo($idCampo, $valorReferente);
                    $dataContainer->setNuevoValorCampo($idCampo, $valorReferente); */
                    
		    $dataContainer->setFieldValue($idCampo, $valorReferente, 'both');                    
                    
                    $dataContainer->getcampo($idCampo)->valorOrig = $valorReferente;
                }
            }
            unset($valorReferente);
        }

        if (isset($this->parameters[$idCampo])) {
            /*
              if (isset($this->parameters['autoprint']) && $this->parameters['autoprint'] != '')
              $valorGet = urldecode($this->parameters[$idCampo]);
              else
              $valorGet = stripslashes(stripslashes($this->parameters[$idCampo]));
            */
            $valorGet = urldecode($this->parameters[$idCampo]);

            if ($refField->TipoDato == 'date' && $valorGet != '') {
                $fecha = date('d/m/Y', strtotime($valorGet));
                $valorGet = $fecha;
            }
            $valorGet2 = $valorGet;
            if ($refField->TipoDato == 'varchar') {
                $valorGet2 = "'" . $valorGet . "'";
            }

            // ver como hacer con parametros GET que no sean con operador igual
            $operador = '';
            $reemplazo = '';
            $operador = (string) $campo->operador;
            if ($operador == '') {
                $operador = '=';
                $reemplazo = 'reemplazo';
            }
            if (isset($this->parameters['__OP__' . $idCampo])) {
                $operador = $this->parameters['__OP__' . $idCampo];
                //echo ' | '.$operador.' | ';
            }
            if ($operador == 'like')
                $valorGet2 = "'%" . $valorGet . "%'";
            $dataContainer->addCondicion($idCampo, $operador, $valorGet2, 'and', $reemplazo, 'false');
            $dataContainer->setCampo($idCampo, $valorGet);
            //	$dataContainer->getcampo($idCampo)->valorOrig;
            $refField->setValorOriginal($valorGet);

            $dataContainer->setNuevoValorCampo($idCampo, $valorGet);
        }

        // Leo las Condiciones
        foreach ($campo->condicion as $condicion) {
            $logicOperator = (string) $condicion['operador'];
            if ($logicOperator == '')
                $logicOperator = '=';
            if ($logicOperator == '=')
                $opc = 'reemplazo';
            else
                $opc = '';

            $valorCondicion = (string) $campo->condicion;
            $valCond2 = '';
            
            if ((string) $condicion['valorde'] != '' && $dataContainerPadre != null) {
                $valorde = (string) $condicion['valorde'];
                $parentField = $dataContainerPadre->getCampo($valorde);
                $valCond2 = $parentField->valor;                    
                
                if (is_array($parentField->opcion) && $valCond2 == ''){
                    $valCond2 = current( current($parentField->opcion));                    
                }
                


//                			loger( 'Padre valor de'.$valorde.'='.$valCond2, 'valorde');
                $valorCondicion = $valCond2;

                if (strpos($refField->TipoDato, 'char') !== false)
                    $valorCondicion = "'" . $valCond2 . "'";
            }
            else
            if ((string) $condicion['valorde'] != '') {

                $valorde = (string) $condicion['valorde'];
                $field = $dataContainer->getCampo($valorde);
                if (is_object($field)) {
                    $valorCondicion = (isset($field->valor)) ? $field->valor : '';
                    if (strpos($refField->TipoDato, 'char') !== false)
                        $valorCondicion = "'" . $valorCondicion . "'";
                }
            }
            $opLogico = 'and';
            if ((string) $condicion['logic'] != '')
                $opLogico = (string) $condicion['logic'];
            $dataContainer->addCondicion($idCampo, $logicOperator, $valorCondicion, $opLogico, $opc, false, null, (string) $condicion['fixed']);
        }


        // Container Variable Setter
        foreach ($campo->setvar as $variable) {
            $varName = (string) $variable;
            $refField->setters[] = array('varname' => $varName);
        }

        foreach ($campo->valor as $val) {
            if ((string) $val['eval'] == 'true') {
                if ((string) $val['retval'] != '') {
                    $code = $val;
                    $code .= '$val = ' . (string) $val['retval'] . ' ;';
                    eval($code);
                } else {
		    $val = $this->replaceFieldValues($dataContainer, $val);
                    $val = '$val = ' . $val . ' ;';
                    eval($val);
                }
                
                if (is_object($val) 
                //&& ($val instance_of Closure)
                ){
                      $val = $val();
                }
                
                
            }
            if (is_array($val) && isset($val['default'] ) && (string) $val['default'] == 'true') {
                $refField->defaultValue = (string) $val;
            }

            $refField->setValor((string) $val);
            $refField->setValorOriginal((string) $val);
        }

        if (isset($refField->rememberValue) && $refField->rememberValue == "true"){
            $fieldId = 'FIELD::'.$dataContainer->xml.'_'.$refField->NombreCampo;
            $htxParameter = new Histrix_Parameter($fieldId, $refField->Etiqueta, $refField->value, $_SESSION['login']);
            $refField->valor = $htxParameter->getValue();

        }

        // Container Variable Getter
        foreach ($campo->getvar as $variable) {
            $varName = (string) $variable;
            $val = $dataContainer->{$varName};

            $refField->setValor($val);
            $refField->setValorOriginal($val);
        }


        foreach ($campo->complete as $comp) {
            $refField->onComplete((string) $comp);
        }

        foreach ($campo->parametro as $par) {
            //$dataContainerParametro = leoXML($dirXML, (string) $par['id'], true, null, $subdir);
            $parxml = ($par['xml'] != '')? (string) $par['xml']: (string) $par['id'];

            if ($parxml != ''){
                $xmlReader = new Histrix_XmlReader($dirXML, $parxml, true, $dataContainer->xml, $subdir);
                $xmlReader->addParameters($this->parameters);
                $xmlReader->serialize = false;

                $dataContainerParametro = $xmlReader->getContainer();

            } else {
                $dataContainerParametro = new ContDatos($tablename, 'param', 'consulta');

            }

            $dataContainerParametro->retorna = (string) $par['retorna'];

            $parameterField = ($par->field) ? $par->field : $par->campo; // replace old tag
            foreach ($parameterField as $campopar) {
                $this->loadContainer($dataContainerParametro, $campopar, '', $dirXML, $dataContainer, $subdir);
            }
            $dataContainerParametro->cargoCampos();
            $valor = $dataContainerParametro->getCampo($dataContainerParametro->retorna)->getValor();
            $refField->setValor($valor);
            $refField->setValorOriginal($valor);
            $refField->addContenedorParametro($dataContainerParametro);
            //     $this->serializados[$dataContainer->xml.'_'.$dataContainerParametro->xml] = $dataContainer->xml.'_'.$dataContainerParametro->xml;
        }


        // Incorpor en el objeto registroEx del Contenedor la relacion con otro contenedor
        foreach ($campo->registro as $registro) {

            $id = (string) $registro['id'];
            $xmlReg = (string) $registro['xml'];
            if ($dataContainer->registrosEx[$id] == '')
                $dataContainer->registrosEx[$id] = new RegistroEx($id, $xmlReg);
            foreach ($registro->destino as $destReg)
                $dataContainer->registrosEx[$id]->par[$idCampo] = (string) $destReg;
        }
        $ifs = $campo->if;
        if ($campo->IF != '') {
            $this->fixXml = true;
            $ifs = $campo->IF;
        }

        foreach ($ifs as $if) {
            $refField->ifs = new Condicion((string) $if['exp'], (string) $if['operador'], (string) $if['valor'], '', '');

            $ifTrue = (isset($if->verdadero)) ? (string) $if->verdadero : (string) $if->true;
            $refField->ifs->verdadero = $ifTrue;

            $ifFalse = (isset($if->falso)) ? (string) $if->falso : (string) $if->false;
            $refField->ifs->falso = $ifFalse;

            if (isset($if->verdadero))
                $this->fixXml = true;
        }

        // cargo un contenedor en un campo

        if ((string) $campo['obj'] != '') {

            // TODO CHANGED
            $customSubdir = $subdir;
            if (isset($campo['objdir'])) {

                $objdir = (string) $campo['objdir'];
                if (substr($objdir, 0, 1) == '/')
                    $customSubdir = $objdir;
                else
                    $customSubdir .= '/' . $objdir;
            }
            //$subdir = (isset($campo['objdir']))? (string) $campo['objdir']:$subdir;
            //
            //          $xmlOrig = (isset($dataContainer->xmlOrig))?$dataContainer->xmlOrig:'';
            $xmlOrig = $dataContainer->xml;

            $xmlReader = new Histrix_XmlReader($dirXML, (string) $campo['obj'], $sub = false, $xmlOrig, $customSubdir, true);
            $xmlReader->addParameters($this->parameters);
            $xmlReader->addReferentContainer($this->ContenedorReferente);

            $objdet = $xmlReader->getContainer();
            $objdet->xmlpadre = $dataContainer->xml;
            $objdet->parentInstance = $dataContainer->getInstance();

            $objdet->xmlReferente = $dataContainer->xml;
            $objdet->isInner = 'true';

            // Propagate Loggin and Notification
            if (isset($dataContainer->_menuId))
                $objdet->_menuId = $dataContainer->_menuId;
            if (isset($dataContainer->log))
                $objdet->log = $dataContainer->log;

            //$refField->contExterno = $objdet;
            if (isset($this->parameters[$idCampo])) {

                $valorGet = $this->parameters[$idCampo];
                if ($refField->TipoDato == 'date') {
                    $fecha = date('d/m/Y', strtotime($valorGet));
                    $valorGet = $fecha;
                }
                // 	ver como hacer con parametros GET que no sean con operador igual
                $objdet->addCondicion($idCampo, '=', $valorGet, 'and', 'reemplazo', 'false');
                $objdet->setCampo($idCampo, $valorGet);
                $objdet->setNuevoValorCampo($idCampo, $valorGet);
                $objdet->getCampo($idCampo)->setValorOriginal($valorGet);

            }

            $refField->esTabla = true;
            $refField->showObjTabla = (string) $campo['showObjTabla'];
            $refField->showValor = (string) $campo['showValor'];
            $refField->contExterno = $objdet;

            if ($this->serialize) {
                $serializedId = Histrix_XmlReader::serializeContainer($objdet);
                /*
                if ($serializedId != '')
                    $this->serializados[$serializedId] = $serializedId;
                    */
            }
        }

        // Cargo los nombres de los campos que la actualizacion de este campo disparará



        foreach ($campo->actualiza as $actualiza) {

            $updateField = ($actualiza->field) ? $actualiza->field : $actualiza->campo; //replace old tag
            foreach ($updateField as $camporef) {
                $id = (string) $camporef['id'];
                $destino = (string) $camporef['destino'];

                $refField->actualiza[$id] = $id;    // Campo que Actualiza
                $refField->destino[] = $destino;  // Campo del cont Int del campo de destino

                $refField->actualizarCampo[$id] = $id;
                $refField->actualizarFilter[$id] = (string) $camporef['filter'];
                ;
                $refField->actualizarCampoXml[$id] = (string) $camporef['xml'];
                ;
                $refField->actualizarDestino[$id][$destino] = $destino;
            }
        }

        foreach ($campo->attribute as $attrib) {
            $refField->valAttribute[(string) $attrib['id']] = (string) $attrib;
            $attrSource = $dataContainer->getCampo((string) $attrib);

            $valAtt = (isset($attrSource->valor)) ? $attrSource->valor : '';
            if ((string) $attrib['id'] == 'oculto')
                $refField->setOculto((string) $valAtt);

            $refField->{(string) $attrib['id']} = (string) $valAtt;
            $refField->Parametro[(string) $attrib['id']] = (string) $valAtt;
        }


        /**
         * Options Loading  // OLD METHOD DEPRECATED
         */
        $fieldOptions = isset($campo->options) ? $campo->options : $campo->opciones;
        if (isset($campo->opciones))
            $this->fixXml = true;
        foreach ($fieldOptions as $opciones) {
            $fieldOption = (isset($opciones->option)) ? $opciones->option : $opciones->opcion;
            foreach ($fieldOption as $opcion)
                $dataContainer->addOpcion((string) $campo['id'], (string) $opcion['valor'], $this->translator->translate((string) $opcion));
        }


        /*
         * CARGO LAS REFERENCIAS EXTERNAS DE OTRAS TABLAS
        */

        $showKey = 'false';
        foreach ($campo->tabla_ex as $key => $valor) {
            $helpdir = $subdir;
            if ((string) $campo->tabla_ex["xml"] != '') {
                if (isset($campo->tabla_ex["dir"])) {
                    $objdir = (string) $campo->tabla_ex["dir"];
                    if (substr($objdir, 0, 1) == '/')
                        $helpdir = $objdir;
                    else
                        $helpdir = $subdir . '/' . $objdir;
                }

                $xmlReader = new Histrix_XmlReader($dirXML, (string) $campo->tabla_ex["xml"], true, $dataContainer->xml, $helpdir);
                $xmlReader->serialize = false;
                $xmlReader->addParameters($this->parameters);
                $micont = $xmlReader->getContainer();
                $micont->xml = (string) $campo->tabla_ex["xml"];
            }
            else
                $micont = new ContDatos((string) $campo->tabla_ex["id"], '', 'consulta');

            $micont->isTable_ex = true;
            $micont->xmlOrig = $dataContainer->xml;


            if (isset($campo->tabla_ex["icono"]))
                $micont->CampoIcono = (string) $campo->tabla_ex["icono"];

            $showKey = 'false';

            $fieldsEx = ($valor->field) ? $valor->field : $valor->campo; // replace old tag
            foreach ($fieldsEx as $fieldEx) {
                $this->loadContainer($micont, $fieldEx, '', $dirXML, $dataContainer, $helpdir);

                if ((string) $fieldEx['muestra'] != '')
                    $showKey = (string) $fieldEx['muestra'];
            }

            $newOrderTag = $valor->order;
            if ($valor->orden != '') {
                $this->fixXml = true;
                $newOrderTag = $valor->orden;
            }

            foreach ($newOrderTag as $orden) {
                $orderField = ($orden->field) ? $orden->field : $orden->campo; // replace old tag
                foreach ($orderField as $campoOrden) {
                    $orderType = isset($campoOrden['orderType'])?$campoOrden['orderType']:$campoOrden['tipoOrden'];                    
                    $micont->setOrden((string) $campoOrden['id'], (string) $campoOrden['tabla'], (string) $orderType);
                }
            }
            $micont->selectVacio = (string) $campo->tabla_ex["vacio"];

            $refField->selectExpand = (string) $campo->tabla_ex["expand"];
            $refField->addContenedor($micont, $showKey);
            $refField->isSelect = 'true';
        }





        /*
         * CARGO LAS AYUDAS CONTEXTUALES QUE APUNTAN A OTRAS TABLAS
         *    DEPRECATED
        */

        foreach ($campo->ayuda_ex as $key => $valor) {
            $xmlH = (string) $campo->ayuda_ex["xml"];
            $idH = (string) $campo->ayuda_ex["id"];

            $subdirhelp = $subdir;
            if (isset($campo->ayuda_ex["dir"])) {

                $objdir = (string) $campo->ayuda_ex["dir"];
                if (substr($objdir, 0, 1) == '/')
                    $subdirhelp = $objdir;
                else
                    $subdirhelp .= '/' . $objdir;
            }

            $xmlid = uniqid('xml', true);
            if ((string) $campo->ayuda_ex["xml"] != '') {
                $xmlReader = new Histrix_XmlReader($dirXML, $xmlH, true, $dataContainer->xml, $subdirhelp, true);
                $xmlReader->addParameters($this->parameters);
                $contHelp = $xmlReader->getContainer();
            }
            else
                $contHelp = new ContDatos($idH, 'Ayuda ' . $campo->Etiqueta, 'ayuda');


            //						$contHelp->xml = ($xml != '')?$xmlH:$idH;
            $contHelp->xml = ($xmlH != '') ? $xmlH : $xmlid;
            /*
            $this->serializados[$dataContainer->xml . '_' . $contHelp->xml] = $dataContainer->xml . '_' . $contHelp->xml;
            $this->serializados['_' . $contHelp->xml] = '_' . $contHelp->xml;
            */
            $helpFields = ($valor->field) ? $valor->field : $valor->campo; // replace old tag

            foreach ($helpFields as $fieldEx)
                $this->loadContainer($contHelp, $fieldEx, '', $dirXML, $dataContainer, $subdir);

            $newOrderTag = $valor->order;
            if ($valor->orden != '') {
                $this->fixXml = true;
                $newOrderTag = $valor->orden;
            }
            foreach ($newOrderTag as $orden) {
                $orderFields = ($orden->field) ? $orden->field : $orden->campo; // replace old tag
                foreach ($orderFields as $campoOrden) {
                    $orderType = isset($campoOrden['orderType'])?$campoOrden['orderType']:$campoOrden['tipoOrden'];
                    $contHelp->setOrden((string) $campoOrden['id'], (string) $campoOrden['tabla'], (string) $orderType);
                }
            }
            $claveAyuda = false;
            foreach ($valor->clave as $key2 => $valor2) {
                $claveAyuda = true;
                $refField->ClaveAyuda = (string) $valor2;
            }
            // Si no tengo La clave ingresada tomo los dos primeros campos de la ayuda
            if ($claveAyuda == false) {
                $j = 0;
                foreach ($contHelp->tablas[$contHelp->TablaBase]->campos as $itemsHlp) {
                    if ($itemsHlp->esOculto())
                        continue;
                    $j++;
                    if ($j == 1)
                        $refField->ClaveAyuda = $itemsHlp->NombreCampo;
                    if ($j == 2)
                        $refField->DescripAyuda = $itemsHlp->NombreCampo;
                }
            }

            foreach ($valor->descripcion as $key3 => $valor3) {
                $refField->DescripAyuda = (string) $valor3;
            }

            $refField->addContenedorAyuda($contHelp);
        }


        /**
         * Add Field Helpers
         */

        foreach ($campo->helper as $helperTag) {
    //        if (isset($campo->helper)) {
      //      $helperTag      = $campo->helper;
            $helperType     = (string) $helperTag['type'];
            $helperXml      = (string) $helperTag['xml'];
            $helperDir      = (string) $helperTag['dir'];
            $helperHtmlTag  = (string) $helperTag['tag'];
            $helperId       = (string) $helperTag['id'];
            $helperIcon     = (string) $helperTag['icon'];
            $helperTitle    = (string) $helperTag['title'];          
            $helperMultiple = (string) $helperTag['multiple'];          
            $helperWidth    = (string) $helperTag['width'];
            $helperHeight   = (string) $helperTag['height'];
            $helperReposition = (string) $helperTag['reposition'];
            $helperModal    = (string) $helperTag['modal'];

            $xmlid = uniqid('xml', true);

            // Force dir
            /*
            if (dirname($helperXml) != '' && dirname($helperXml) != '.'){
                $helperDir = dirname($helperXml);
                $helperXml = basename($helperXml);
            }
            */
            $dirfile = dirfile($helperXml, $helperDir );
            $helperDir = $dirfile['dir'];
            $helperXml = $dirfile['file'];

            // Add relative path
            $newHelperDir = $subdir;

            if ($helperDir != '') {
                if (substr($helperDir, 0, 1) == '/')
                    $newHelperDir = $helperDir;
                else
                    $newHelperDir = $subdir . '/' . $helperDir;
            }



            switch ($helperType) {
                case 'parameter':
                    $helperLabel = (string) $helperTag->label;
                    $helperDefaultValue = (string) $helperTag->defaultValue;
                    $htxParameter = new Histrix_Parameter($helperId, $helperLabel, $helperDefaultValue);
                    $refField->valor = $htxParameter->getValue();
                break;
                case 'combo':
                case 'comboex':
                    foreach ($helperTag->option as $option) {
                        $dataContainer->addOpcion((string) $campo['id'], (string) $option['value'], $this->translator->translate( $option));
                    }

                    if ($helperXml != '' || $helperId != '') {
                        if ($helperXml != '') {
                            $xmlReader = new Histrix_XmlReader($dirXML, $helperXml, true, $dataContainer->xml, $newHelperDir, true);
                            $xmlReader->serialize = false;
                            $xmlReader->addParameters($this->parameters);
                            $micont = $xmlReader->getContainer();
                            $micont->xml = $helperXml;
                        } else {

                            $micont = new ContDatos($helperId, '', 'consulta');
                        }

                        $micont->isTable_ex = true;
                        $micont->xmlOrig = $dataContainer->xml;
                        $micont->parentInstance = $dataContainer->getInstance();

                        if ($helperIcon != '')
                            $micont->CampoIcono = $helperIcon;

                        $showKey = 'false';
                        // load Fields
                        foreach ($helperTag->field as $fieldTag) {
                            $this->loadContainer($micont, $fieldTag, '', $dirXML, $dataContainer, $newHelperDir);
                            if ((string) $fieldTag['showKey'] != '')
                                $showKey = (string) $fieldTag['showKey'];
                        }

                        foreach ($helperTag->order as $orderTag) {
                            foreach ($orderTag as $orderField) {
                                $orderType = isset($campoOrden['orderType'])?$campoOrden['orderType']:$campoOrden['tipoOrden'];                                
                                $micont->setOrden((string) $orderField['id'], (string) $orderField['table'], (string) $orderType);
                            }
                        }

                        $micont->selectVacio = (string) $helperTag["empty"];
                        $refField->selectExpand = (string) $helperTag["expand"];
                        $refField->addContenedor($micont, $showKey);
                        $refField->isSelect = 'true';

                        if ($helperMultiple == 'true')
                            $refField->multiple = 'true';


                    foreach ($helperTag->parameter as $parameter) {
                        $destinationField = (string) $parameter['target'];
                        $refField->paring[$destinationField]['valor'] = (string) $parameter['source'];
                        $refField->paring[$destinationField]['operador'] = (string) $parameter['operator'];
                        
                    }
  
                    }
                    break;
                case "external":

                    if ($helperXml != '') {
                        $xmlReader = new Histrix_XmlReader($dirXML, $helperXml, true, $dataContainer->xml, $newHelperDir, true);
                        $xmlReader->addParameters($this->parameters);
                        $micont = $xmlReader->getContainer();
                        $micont->xml = $helperXml;
                    } else {

                        $micont = new ContDatos($helperId, 'Ayuda' . $campo->Etiqueta, 'ayuda');
                        $micont->xml = $xmlid;
                    }
                    /*
                    $this->serializados[$dataContainer->xml . '_' . $micont->xml] = $dataContainer->xml . '_' . $micont->xml;
                    $this->serializados['_' . $micont->xml] = '_' . $micont->xml;
                    */
                    foreach ($helperTag->field as $fieldTag) {
                        $this->loadContainer($micont, $fieldTag, '', $dirXML, null, $newHelperDir);
                    }

                    foreach ($helperTag->order as $orderTag) {
                        foreach ($orderTag as $orderField) {
                            $orderType = isset($campoOrden['orderType'])?$campoOrden['orderType']:$campoOrden['tipoOrden'];
                            $micont->setOrden((string) $orderField['id'], (string) $orderField['table'], (string) $orderType);
                        }
                    }


                    $claveAyuda = false;
                    foreach ($helperTag->key as $key2 => $valor2) {
                        $claveAyuda = true;
                        $refField->ClaveAyuda = (string) $valor2;
                    }
                    // Si no tengo La clave ingresada tomo los dos primeros campos de la ayuda
                    if ($claveAyuda == false) {
                        $j = 0;
                        foreach ($micont->tablas[$micont->TablaBase]->campos as $itemsHlp) {
                            if ($itemsHlp->esOculto())
                                continue;
                            $j++;
                            if ($j == 1)
                                $refField->ClaveAyuda = $itemsHlp->NombreCampo;
                            if ($j == 2)
                                $refField->DescripAyuda = $itemsHlp->NombreCampo;
                        }
                    }

                    foreach ($helperTag->value as $key3 => $valor3) {
                        $refField->DescripAyuda = (string) $valor3;
                    }

                    $refField->addContenedorAyuda($micont);


                    break;
                case 'object':
                    $xmlOrig = (isset($dataContainer->xmlOrig)) ? $dataContainer->xmlOrig : '';
                    $xmlReader = new Histrix_XmlReader($dirXML, $helperXml, false, $xmlOrig, $newHelperDir, true);
                    $xmlReader->addParameters($this->parameters);
                    $xmlReader->addReferentContainer($this->ContenedorReferente);

                    $objdet = $xmlReader->getContainer();
                    $objdet->xmlpadre = $dataContainer->xml;
                    $objdet->xmlReferente = $dataContainer->xml;
                    $objdet->isInner = 'true';
	                $objdet->parentInstance = $dataContainer->getInstance();

                    // Propagate Loggin and Notification
                    if (isset($dataContainer->_menuId))
                        $objdet->_menuId = $dataContainer->_menuId;
                    if (isset($dataContainer->log))
                        $objdet->log = $dataContainer->log;

                    // cascade value assignation
                    // not sure if this it actually works....
                    if (isset($this->parameters[$idCampo])) {

                        $valorGet = $this->parameters[$idCampo];
                        if ($refField->TipoDato == 'date') {
                            $fecha = date('d/m/Y', strtotime($valorGet));
                            $valorGet = $fecha;
                        }
                        // 	ver como hacer con parametros GET que no sean con operador igual
                        $objdet->addCondicion($idCampo, '=', $valorGet, 'and', 'reemplazo', 'false');
                        $objdet->setCampo($idCampo, $valorGet);
                        $objdet->setNuevoValorCampo($idCampo, $valorGet);
                	$objdet->getCampo($idCampo)->setValorOriginal($valorGet);
                        
                    }

                    $refField->esTabla = true;
                    $refField->showObjTabla = (string) $campo['showObjTabla'];
                    $refField->showValor = (string) $campo['showValor'];
                    $refField->contExterno = $objdet;

                    if ($this->serialize) {
                        $serializedId = Histrix_XmlReader::serializeContainer($objdet);
                        /*
                        if ($serializedId != '')
                            $this->serializados[$serializedId] = $serializedId;
                            */
                    }

                    foreach ($helperTag->parameter as $parameter) {
                        $destinationField = (string) $parameter['target'];
                        $refField->paring[$destinationField]['valor'] = (string) $parameter['source'];
                        $refField->paring[$destinationField]['operador'] = (string) $parameter['operator'];
                        
                        $refField->paring[$destinationField]['reemplazo'] = (string) $parameter['replace'];
                        
                    }
		    
                    break;
                case 'link':
                    $refField->linkint = $helperXml;
                    if ($helperDir != '')
                        $refField->linkdir = $helperDir;
                     if ($helperHtmlTag != '')
                        $refField->linkTag = $helperHtmlTag;
                    if ($helperTitle != '')
                        $refField->linkdes = $helperTitle;
                    if ($helperWidth != '')
                        $refField->linkWidth = $helperWidth;
                    if ($helperHeight != '')
                        $refField->linkHeight = $helperHeight;
                    if ($helperReposition != '')
                        $refField->linkReposition = $helperReposition;
                    if ($helperModal != '')
                        $refField->linkmodal = $helperModal;
                    if ($helperIcon != '')
                        $refField->linkIcon = $helperIcon;


                    foreach ($helperTag->parameter as $parameter) {
                        $destinationField = (string) $parameter['target'];
                        $refField->paring[$destinationField]['valor'] = (string) $parameter['source'];
                        $refField->paring[$destinationField]['operador'] = (string) $parameter['operator'];
                        
                    }
                    break;
            }
        }
    }

    /**
     * read dashborad specific xml format
     * @param simpleXml $simpleXml
     * @return Dashboard
     */
    function readDashboard($simpleXml) {
        $dashboard = new UI_dashboard();
        foreach ($simpleXml->widget as $widget) {

            $id = (string) $widget['id'];

            $myWidget = new Widget($id);

            foreach ($widget->attributes() as $parameter => $value) {
                $myWidget->{$parameter} = (string) $value;
            }

            $myWidget->title = (string) $widget->title;

            // for external content via http
            $myWidget->url = (string) $widget->url;

            // Static content
            $myWidget->text = (string) $widget->text;

            // iframe content
            $myWidget->iframe = (string) $widget->iframe;

            // Add widget to dashboard
            $dashboard->addWidget($myWidget);
        }
        return $dashboard;
    }

    private function replaceFieldValues($dataContainer, $val){
        $fields = $dataContainer->camposaMostrar();
        foreach ($fields as $fieldName) {
            $fieldValue = $dataContainer->getCampo($fieldName)->getValor();
            $val = str_replace('[__'.$fieldName.'__]', $fieldValue, $val);
        }
	return $val;
    }
}

?>
