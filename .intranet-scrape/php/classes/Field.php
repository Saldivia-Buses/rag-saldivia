<?php
/**
 *   Clase Campo | Field Class
 *
 *   @package Histrix
 *   @author Luis M. Melgratti <luis@estudiogenus.com>
 *   @date    24/10/2005
 *
 **/

class Field 
{
        /*
        var $valorCampo;

        var $Parametro;  // Array de los parametros del campo
    // propiedades

    public $fieldName;	// Nombre del campo
    public $Expresion;		// Expresion asociada (si tiene)
    public $Etiqueta;		// Etiqueta (label)
    public $ayuda;			// Texto de Ayuda Asociado al campo
    public $TipoDato;		// Tipo de dato del campo
    public $Formato;		// Formato (tipo php) de representacion
    public $local; 		// Indica si es un campo local que NO hace operaciones SQL
    public $Oculto;		// Si el campo esta oculto en los ABM
    public $noshow;		// Si es un campo hidden
    public $deshabilitado; // Marca si el campo esta deshabilitado en los ABM
    public $valop; 		// Indica si se almacena el valor del campo o el valor de la opcion asociada;
    
    public $Arbol;	// Para el caso de los arboles determina si el campo es hijo o padre

    public $suma;			// Valor Acumulado del Campo

    // SQL
    public $esClave; 		// si forma parte del indice primario
    public $valorOrig;		// Valor original del campo (antes de ser modificado)
    public $valor;			// valor actual
    public $nuevovalor;	// valor a grabar en el proximo sql
    public $orden; 		// si es parte del orden
    public $tipoOrden; 	// ASC O DESC
    public $TablaPadre; // Tabla a la cual pertenece el campo
    public $noselect;		// Si esta en TRUE no figura el campo en los SELECT (se usa para condiciones solamente)

    // Propiedades visuales
    public $Size; 			// Tamaño en caracteres
    public $tammax; 		// tamaño maximo para los inputs (lo toma del contenido de la seleccion actual)
    public $modpos; 		// Indica que no pase a una nueva linea en el ABM (modificador de posicion "nobr")
    public $colspan; 		// colspan
    public $prefijo;		// Prefijo antes del dato (PDF)
    public $foto; 			// Si se debe representar como una foto
    public $norepeat;		// Establece para consultas con N registros de un indice para no repetir datos de ese campo en las consultas
    public $corte;			// Establece el corte para sumas de subtotales con este campo

        // Style
        var $style;
        var $colstyle;
        var $colclass;
        var $getStyle;
        var $clasefila;
        var $Formstyle;

    public $condes; 		// Etiqueta que se llena si el campo clave es correcto
    public $campoTemp; 	// ing. de Movimientos: Nombre del campo relacionado de la Tabla temporal que graba en este campo
    public $campoCab; 		// ing. de Movimientos: Nombre del campo relacionado de la Tabla CABECERA que graba en este campo
    public $campoPadre;	// ing. de Movimientos: Nombre del campo relacionado de la Tabla Padre que graba en este campo

    public $ifs; 			// ing. de Movimientos: condiciones para autocalculo de campos
    public $copia; 		// nombre del campo donde se copia el valor
    public $refresco; 		// Si El campo es un chech box y esta en TRUE se refresca el calculo con cada marca/desmarca

    public $Detalle;	// array de nombres de los campo con el que se relaciona en el detalle

        var $noEmpty;
    // Array con Opciones
    public $opcion;
    public $Condiciones;	// Array con Condiciones de este campo para el Select

    // Contenedor con Datos Externos
    public $contExterno;
    public $esTabla; 		//si el contenedor representa una tabla de detalle interno

    // Datos para la ayuda contextual
    public $contAyuda; // Contenedor Ayuda Externa
    //var $TablaAyuda;
    public $ClaveAyuda; // nombre del campo relacionado

    // Parametros
    public $ContParametro; // Contenedor que tiene el parametro (ej. Numerador) asociado
    public $onComplete; 	// valor del campo cuando se completa la transaccion (uso en numeradores)
    public $param; 		// indica que el contenido de este campo es condicion del parametro de otro

    // PDF
    public $PDFnolabel;	// si corresponde la etiqueta en el PDF
    public $PDFsize;		// tamaño de la fuente en PDF
        var $PDFanchofijo;
    public $posX;			// Posicion en el formulario PDF
    public $posY;  		// Posicion en el formulario PDF

        var $firstnt;

        // Otros
        // TODO DOCUMENT
        var $font;
        var $forceSize;
        var $maxSize;
        var $activa;
        var $color;
        var $noZero;
        var $fieldTypeype;
        var $tooltip;
        var $importacion;
        var $valauto;
        var $conLabel;
        var $sincelda;
        var $maxsize;

        var $valAttribute;
        var $paring;
        var $jsfunction;
        var $jsevaltotal;
        var $jseval;
        var $jsextract;
        var $jsExec;
        var $tdWidth;

        var $aletras;
        var $linkint;
        var $validar;
        var $cuit;
        var $onformchange;
        var $required;
        var $min;
        var $max;
        var $autoing;
        var $pintado;
        var $habilita;
        var $conditionalDisplay;
        var $Tammax;
        var $lastTabindex;
        var $activador;
        var $linkExt;
        var $linkdir;
        var $linkPrint;
        var $linktab;
        var $columnLabel;

        var $actualizarCampo;
        var $lastValue;

        var $ignoreOption;
        var $tablaOrden;
        var $idlbl;
        var $dateselector;

        var $refresh;
        var $noupdate;
        var $searchRecord;
        var $autocompletar;
        var $showInForm;
        var $autofield;
        var $autoinc;
        var $idName;
        var $selectExpand;
        var $checkType;
        var $innerTable;
        var $valoropcion;

        */
/**
* Constructor Object Campo
*

 * @param string $Nom Field Name
 * @param string $Exp Field SQL expression
 * @param string $Etiq  Field Label
 * @param string $formato Field Format
 */
    public function __construct($fieldName, $expression='', $label='', $format='')
    {
        $this->NombreCampo 	= $fieldName;
        if ($expression != '')
            $this->Expresion 	= $expression;
        if ($label != '')
            $this->Etiqueta 	= $label;
        if ($format != '')
            $this->Formato 		= $format;
        /* tamaño y tipo por defecto */
        $this->Size 		= 12;
        $this->TipoDato 	= 'varchar';
        $this->Suma= 0;
        $this->uid 	= UID::getUID($fieldName, true);

    }

    /**
      Agrega el contenedor de los Parametros
     */
    public function addContenedorParametro($objParametro)
    {
        $this->ContParametro = $objParametro;
    }

    /**
     * Add Help DataContainer to Field
     * @param ContDatos $conte Help DataContainer
     */
    public function addContenedorAyuda($conte)
    {
        $this->contAyuda = $conte;
    }

    /**
      get Help Container
      @return ContDatos
     */
    public function getContenedorAyuda()
    {
        if (isset($this->contAyuda))
            return $this->contAyuda;
    else return false;
    }

    /**
      get Help Container SQL select statement
      @return string
     */
    public function getSelectAyuda()
    {
        return $this->contAyuda->getSelect();
    }
    /**
      Agrego un contenedor para las tablas externas
     */
    public function addContenedor($conte, $muestraClave = 'false')
    {
        $this->contExterno = $conte;

        $this->llenoOpciones($muestraClave);

    }

   /**
     Fill Options from database Resultset
    * 
    * @param string  $muestraClave true / false - shows keys in string result
    * @param integer $row          Row NUmber
    * 
    * @return none
    */
    public function llenoOpciones($muestraClave = 'false', $row='')
    {
        $recordSet  = $this->contExterno->Select();
        $rowIndex   = 0;
        $campos     = _num_fields($recordSet);

        // add Empty option
        if ($this->contExterno->selectVacio=='true')
            $this->addOpcion('', '');

        while ($row = _fetch_array($recordSet)) {
            $fieldIndex = 0;
            $rowIndex++;

            while ($fieldIndex < $campos) {
                $fieldIndex++;

                // store metadata just in first row
                if ($rowIndex == 1) {
                    $campoactual[$fieldIndex] = _field_name($recordSet, $fieldIndex);
                    $hidden[$fieldIndex] = false;
                    if (isset($campoactual[$fieldIndex])) {
                        $field = $this->contExterno->getCampo($campoactual[$fieldIndex]);
                        if (isset($field->oculto) && $field->oculto == "true")
                            $hidden[$fieldIndex] = true;
                    }
                }

                if ($hidden[$fieldIndex])  continue;

                if ($fieldIndex > 1)
                    $valores[$campoactual[$fieldIndex]] = $row[$campoactual[$fieldIndex]];

            }
            $clave = $row[$campoactual[1]];
            $valor = $row[$campoactual[2]];
            if ($muestraClave == 'true')
                $valor = $clave.' - '.$valor;

            $this->addOpcion($clave, $valor, $row);
            $this->addOpcion($clave, $valores, $row);
        }

        if ($this->contExterno->selectVacio == 'true') {
            if (isset($valores)) {
                foreach ($valores as $key => $val) {
                    $empty[$key] = '';
                }
            }
            $this->addOpcion('', $empty);
        }

    }

   /**
    * addOption to combo so select inputs

    * @param string $key   option key
    * @param string $value option Description
    * @param string $row   row data
    * @return void
    */
    public function addOpcion($key, $value,$row='')
    {
        /*
        if ($row != '')
           $this->rowOption[$key] = $value;
        */
        $this->opcion[$key] = $value;
    }

    /**
      Agrega una condicion logica en el campo
     *
     */
    public function addCondicion($logic, $oper, $val, $modificador = '', $premod = '', $controloTipo='true', $grupo = '', $fixed=false)
    {
        if (isset($this->local) && $this->local === true)
            return false;
        $this->xsdType = Types::getTypeXSD($this->TipoDato, 'xsd:integer');

        $label = '';
        if ($this->TipoDato == 'date') {
//            $fec = Types::removeQuotes(
            $fec = $this->getValorGen($val, $this->TipoDato);
  //           die($val);
            $arr=explode("-",$fec); // splitting the array
            $year=$arr[0]; // first element of the array is month
            $month=$arr[1]; // second element is date
            $day=$arr[2]; // third element is year
            if (is_numeric($year) && is_numeric($month) && is_numeric($day)) {
                if (checkdate($month,$day,$year)) {

              //  echo $month, $day, $year;
                    $val = "'".$fec."'";
                }
            }
        }

        if ($this->TipoDato == 'time') {
//            $fec = Types::removeQuotes(
            $fec = $this->getValorGen($val, $this->TipoDato);
  //           die($val);
            $arr=explode(":",$fec); // splitting the array

            if (is_numeric($arr[0]) &&  // Hours
                is_numeric($arr[1]) &&  // Minutes
                is_numeric($arr[2])) {  // Seconds
                    $val = "'".$fec."'";
            }
        }


        // Verifico el tipo del dato con el valor
        if ($controloTipo =='true' && !is_array($val)) {
            switch ($this->xsdType) {
                case "xsd:decimal" :
                    if (!(is_numeric($val))) {

                        return false;
                    }
                    break;
                case "xsd:integer" :
                    if (!(is_numeric($val))) {

                        return false;
                    }
                    break;
            }
        }
        if (isset($this->fullSearch) && $this->fullSearch == 'true') {
            $val = str_replace(' ', '%', $val );
        }

        $hay = false;
        if ($modificador == 'reemplazo') {
        //Reemplazo una condicion existente
            if (isset ($this->Condiciones))
                foreach ($this->Condiciones as $key => $valor) {
                    if (trim($valor->operador) == trim($oper)) {
                        $this->Condiciones[$key] = new Condicion($logic, $oper, $val, $label, $premod, $grupo, $fixed);
                        $this->Condiciones[$key]->tipo = $this->TipoDato;
                        $hay = true;
                    }
                }
        }
        if ($hay == false) {
            $cond  = new Condicion($logic, $oper, $val, $label, $premod, $grupo, $fixed);
            $cond->tipo = $this->TipoDato;
            $key = trim($oper.$val);
            $this->Condiciones[$key] = $cond;
        }

        return true;
    }

    /**
      Borra las condiciones del campo
     */
    public function delCondiciones()
    {
    // vacio las condiciones
        if($this->Condiciones != '')
            foreach ($this->Condiciones as  $ncod => $cond) {
                if (!$cond->fixed) unset($this->Condiciones[$ncod]);

            }
    }

    public function getFormatedValue($value = NULL)
    {
        if (is_null($value)) $value = $this->valor;

        if ($this->Formato != '' && $value != '') {
            if ($this->TipoDato =='date' || $this->TipoDato =='time') {
                $value = date($this->Formato, strtotime($value));
            } else {

                $value = sprintf($this->Formato, $value);
                //die();
            }
        }

        return $value;

    }

    public static function getValorGen($val, $tipo)
    {
        if ($tipo == 'date' && $val != '') {
            if (strpos($val, '__]')) {
            // dejo el valor
                return $val;
            }
            //remove quotes
            $val= str_replace('"','' , $val );
            $val = trim($val);
            $ffecha = $_SESSION['ffecha'];

            /* Me fijo donde esta el primer guion */

            if (substr($val, 4, 1) =='-' || substr($val, 4, 1) =='/' ) {

            } else {

                $ffecha = str_replace("dd", substr($val, 0, 2), $ffecha );
                $ffecha = str_replace("mm", substr($val, 3, 2), $ffecha );
                $val	= str_replace("yyyy", substr($val, 6, 4), $ffecha );
            }
        }
        return $val;
    }

    /**
      Devuelve el valor del campo
     */
    public function getValor()
    {
        $value = (isset($this->valor))?$this->valor:'';

        return $this->getValorGen($value, $this->TipoDato);
    }

    /**
      Devuelve el nuevo valor del campo
     */
    public function getNuevoValor()
    {
        return $this->getValorGen($this->nuevovalor, $this->TipoDato);
    }

    /**
     Establece un valor para el campo
     */
    public function setValor($valor)
    {
        $this->valor = Types::checkToday($valor,  $this->TipoDato);
        $this->updateSetters($this->valor);
    }

    /**
    set last value on field
    */
    public function setLastValue($value)
    {
        $this->ultimo= $value;
        $this->updateSetters($value);
    }

    /**
     Set Container variable value with field value
     */
    public function updateSetters($value)
    {
        if (isset($this->setters)) {
            foreach ($this->setters as $setter) {
                $varName = $setter['varname'];
                $this->_DataContainerRef->{$varName} = $value;
            }
        }

        // set remembered states in htxoption table
        // 
        if (isset($this->rememberValue) && $this->rememberValue == "true"){
            $fieldId = 'FIELD::'.$this->_DataContainerRef->xml.'_'.$this->NombreCampo;
            $htxParameter = new Histrix_Parameter($fieldId, $this->Etiqueta, $value, $_SESSION['login']);
            $htxParameter->updateData($value);


        }


    }
    /**
      Set attributtes of filed based on an array
      @param Array
     */
    public function setAttributes($row)
    {
        if (isset($this->valAttribute) && $this->valAttribute != '') {
            foreach ($this->valAttribute as $attribID => $attrib) {

          if ($attribID == 'attributes') {

                    $parameters = $row[$attrib];
            parse_str($parameters, $innerAttributes);

            foreach ($innerAttributes as $innerid => $valueInner) {
            if (isset($row[$valueInner]))
                $valAtt = $row[$valueInner];
            else {
                $valAtt = $valueInner;
            }
            if ($valAtt != '') {

                        $this->{$innerid} = (string) $valAtt;
                        $this->Parametro[$innerid]= (string) $valAtt;

            }
            }

          } else {

//                if (isset($row[$attrib])) {
                    $valAtt = $row[$attrib];

                    if ($attribID == 'oculto') {
                        if ($valAtt === 'true' || $valAtt === true)
                            $this->setOculto($valAtt);
                    }

                    $this->{$attribID} = (string) $valAtt;
                    $this->Parametro[$attribID]= (string) $valAtt;
  //              }

        }

            }
        }
    }

    /**
      Establece un valor para el campo cuando se completa la transaccion
     */
    public function onComplete($valor)
    {
        $this->onComplete = Types::checkToday($valor,  $this->TipoDato);
    }

    /**
      Establece el nuevo valor del campo
     */
    public function setNuevoValor($valor)
    {
        $this->nuevovalor = Types::checkToday($valor,  $this->TipoDato);
        $this->updateSetters($this->nuevovalor);
    }

    /**
      Establece un valor Original para el campo
     */
    public function setValorOriginal($valor)
    {
        /* casos especiales*/
        $this->valorOrig = Types::checkToday($valor,  $this->TipoDato);

    }

    /**
      Restaura el valor original del campo
     */
    public function restaurarValores()
    {
        $this->valor = $this->valorOrig;
    }

    /**
      Establece el Tipo de Dato a Almacenar en el campo
     */
    public function setTipoDato($type)
    {
        $this->TipoDato = $type;
        $fieldType = 'FieldType_'.$type;

        if ($type != ''){
            if (class_exists($fieldType)) {
                if (defined($fieldType.'::HIDDEN'))
                    $this->colstyle = 'display:none;';
            } else {
                loger($fieldType, 'fieldtypes_ui.log');
            }
        }
    }

    /**
      Establece el nombre del campo con que se relaciona este en
      Las consultas con detalle
     */
    public function setDetalle($det)
    {
        $this->Detalle[] = $det;
    }

    /**
      Establece la ayuda contextual del campo
     */
    public function setAyuda($hlp)
    {
        $this->ayuda = $hlp;
    }

    /**
      Devuelve la ayuda contextual del campo
     */
    public function getAyuda()
    {
        if (isset($this->ayuda))
            return $this->ayuda;
    }

    /**
     * sumField add current value to acumulated field value and stores it in $uiClass
     *
     * @return void
     * @author
     **/
    private function sumField($value, $uiClass)
    {

            if ($this->TipoDato =='time') {
                // Hora de la suma previa en Segundos
                if (isset($uiClass->Suma[$this->NombreCampo]))
                    $horaOrig = $uiClass->Suma[$this->NombreCampo];
                else $horaOrig='';
                $arrOrig=explode(":",$horaOrig); // splitting the array

                if (count($arrOrig) >= 2) {

                    $hhOrig=$arrOrig[0]; // horas
                    $mmOrig=$arrOrig[1]; // minutos

                    if (isset($arrOrig[2]))
                                $ssOrig=$arrOrig[2]; // segundos
                    else $ssOrig = 0;
                    $segundosOrig= $hhOrig * 3600 + $mmOrig * 60 + $ssOrig;

                } else
                    $segundosOrig = 0;

                $segundosSub = $segundosOrig;

                // Hora actual en segundos
                $arr=explode(":",$value); // splitting the array

                if (count($arr) >= 2) {
                    $hhVal=$arr[0]; // horas
                    $mmVal=$arr[1]; // minutos

                    if (isset($arr[2]))
                                $ssVal=$arr[2]; // segundos

                    $segundos= $hhVal * 3600 + $mmVal * 60 + $ssVal;
                } else $segundos = 0;

                // Suma de segundos
                $sumseg    = $segundos + $segundosOrig;
                $sumsegSub = $segundos + $segundosSub;

                // Total
                $horas=floor($sumseg/3600);
                $minutos=floor(($sumseg%3600)/60);//1
                $segs= $sumseg%60;//1
                $valorNuevo = sprintf("%02d",  $horas).sprintf(":%02d",  $minutos).(($this->seconds=='false')?'':sprintf(":%02d",  $segs));

                // subtotal
                $horasSub=floor($sumsegSub/3600);
                $minutosSub=floor(($sumsegSub%3600)/60);//1
                $segsSub= $sumsegSub%60;//1
                $valorNuevoSub = sprintf("%02d",   $horasSub).sprintf(":%02d",  $minutosSub).(($this->seconds=='false')?'':sprintf(":%02d",  $segsSub));

                $uiClass->Suma[$this->NombreCampo] = $valorNuevo;
                $uiClass->Subtotal[$this->NombreCampo] = $valorNuevoSub;

            } else {
                if (isset($uiClass->Suma[$this->NombreCampo]))
                    $uiClass->Suma[$this->NombreCampo] += $value;
                else $uiClass->Suma[$this->NombreCampo] = $value;

                if (isset($uiClass->Subtotal[$this->NombreCampo]))
                    $uiClass->Subtotal[$this->NombreCampo] += $value;
                else $uiClass->Subtotal[$this->NombreCampo] = $value;
            }
    }

    /**
     * render field as a table cell

     * @param UI      $uiClass  Interface class
     * @param string  $nomcampo field name
     * @param string  $valCampo field value
     * @param integer $orden    row order
     * @param integer $posX     x position
     * @param integer $posY     y position
     * @param string  $params   extra parameters
     * @param string  $tag      tag to use
     *
     * @return string
     */
    public function renderCell( $uiClass,  $nomcampo, $valCampo, $orden, $posX, $posY, $params=null, $tag='td')
    {

        $fieldNameReference = false;
	    $tagParameter = array();

        // field detail
        $detallecampo = $this->getUrlVariableString($valCampo);
        if ($detallecampo != '') $uiClass->det .= '&amp;'.$detallecampo;

        $valfecha = '';
        if ($valCampo == '0000-00-00') {
            $valCampo = '';
        } else
            if ($this->TipoDato == 'date' && $valCampo != '') {
                $valfecha = $valCampo;
                $valCampo = date("d/m/Y", strtotime($valCampo));
            }
        if ($this->TipoDato == 'time')
            $valfecha = $valCampo;

        $valor = $valCampo;

        /////////////////////////////////
        // Chequeo repetidos
        ////////////////////////////////

        $uiClass->norepeat = false;
        if (isset($this->corte) && $this->corte == 'true') {
            if ($uiClass->valcorte != $valor  &&  $posY > 1) {
                // TODO
                // THIS CODE IS UNUSED!!!!
                $salida .= $uiClass->showTotales('sub', $uiClass->valcorte);
            }
            $uiClass->valcorte = $valor;
        }

        if (isset($this->norepeat) && $this->norepeat == 'true') {
            $uiClass->norepeat = true;
            if ($uiClass->camposIndexados != '')
                foreach ($uiClass->camposIndexados as $cindexado) {
                    if ($cindexado['anterior']!= $cindexado['actual'] )
                        $uiClass->norepeat=false;
                }
        }

        ////////////////////////////////////
        // SUM FIELDS values
        ///////////////////////////////////
        $doSum = (isset($params['sum']))?$params['sum']:'';

        if (isset($uiClass->Datos->sumaCampo[$this->NombreCampo]) &&
            $uiClass->norepeat != true && $doSum != 'false' ) {

            $this->sumField($valor, $uiClass);
        }

        if (isset($uiClass->Datos->acumulaCampo[$this->NombreCampo]) && $uiClass->norepeat != true) {
            $valor = $uiClass->Suma[$this->NombreCampo];
        }

        if (isset($uiClass->camposIndexados[$this->NombreCampo])) {
            $uiClass->camposIndexados[$this->NombreCampo]['anterior'] = $uiClass->camposIndexados[$this->NombreCampo]['actual'];
            $uiClass->camposIndexados[$this->NombreCampo]['actual'] = $valor;
        }

        $dataType = Types::getTypeXSD($this->TipoDato);
        if ($dataType == 'xsd:integer'  || $dataType == 'xsd:decimal') {
            if ($valor == '')
                $valor = 0;
            $this->setLastValue($valor);
        }
        // ultimo valor del campo

        if ($valor != '')
            $this->setLastValue($valor);

        // Formato por Defecto
        $valorid = $valor;

        // Format field by class
        $fieldType = 'FieldType_'.$this->TipoDato;

        if ($this->TipoDato != ''){
            if (class_exists($fieldType)){
              //  $Type = new $fieldType($this);
                $alignCons = constant($fieldType.'::ALIGN');
                if ($alignCons != 'left') {

                    $tagParameter['align'] = $alignCons;
                }

                //$arrayAttributes['dir']     = constant($fieldType.'::DIR');

                 // Add custom Parameters
                 //   $modif['custom'] = $fieldType::customCellParameters();

                if (method_exists($fieldType ,'customValue')) {

                    $renderParameters['order'] = $orden;
                    $valor = $fieldType::customValue($valor, $this, $renderParameters);
                }

            } else {

                loger('falta: '.dirname(__FILE__).'/FieldType/'.$fieldType.'.php', 'fieldtypes_ui.log');
            }
        }
        $valCampo = $valor;

        switch ($this->TipoDato) {

            case "check" :
                   $fieldNameReference = true;   /// what is this for????
                break;

            case "button":
                if ($valor != '') {
                    $btnval = new Html_button($valor, '' ,$valor );

                    $btnval->addParameter('class', $this->Class);
                    // add custom Javascript Events
                    if ($this->jsfunction)
                        foreach ($this->jsfunction as $jsevent => $jsfunctions) {
                            foreach($jsfunctions as  $jsfunction)
                                $btnval->addEvent($jsevent, $jsfunction, true); // append function
                        }
                    $valor = $btnval->show();
                }
                break;
        }

        $imgIcon = '';

        // Si tiene opciones de un combo
        $valop = (isset($this->valop)) ? $this->valop:'';
        if (isset($this->opcion) && count($this->opcion) > 0 && $this->TipoDato != "check" && $valop != 'true') {

            if (isset($this->opcion[$valCampo]))
                $valor = $this->opcion[$valCampo];

            if (isset($this->ignoreOption) && $this->ignoreOption == 'true') {
                if (!isset($this->opcion[$valCampo]))
                    $valor = $valCampo;
            }

            if (is_array($valor)) $valor = current($valor);

            if (isset($this->contExterno) && isset($this->contExterno->CampoIcono)) {
                $imgIcon = $valor[$this->contExterno->CampoIcono];
                if ($imgIcon != '')
                    $imgIcon ='<img src="../img/'.$imgIcon.'" alt="'.$valor.'" title="'.$valor.'">';
            }

        }

        $formatoCampo = (isset($this->Formato))?$this->Formato:null;

        // Link interno
        if (isset($this->linkint) && $this->linkint != '') {

            $valorid = $valor;
            if ( !(isset($uiClass->fillTextArray)) || !$uiClass->fillTextArray)
                $valor = $uiClass->linkButton($this, $valor, $valfecha, $params, $formatoCampo);

        } else {

            if ($formatoCampo != '') {
                if ($this->TipoDato == 'date') {
                        $valor = $this->getFormatedValue($valfecha);
                        //$size = strlen($valor);

                } else {
                    $valor = $this->getFormatedValue($valor);
                }
            }

        }

        ///////////////////
        // cellMaxLength
        ///////////////////

        $limitSize = (isset($this->cellMaxLength) &&  $this->cellMaxLength != '' && strlen($valor) > $this->cellMaxLength);
        $valor     = ($limitSize)? substr($valor, 0 , $this->cellMaxLength) .' <span forceTittle="true">...'.$uiClass->i18n['more'].'</span>': $valor;

        // TODO: Reprogramar los campos passwods
        if (isset($this->FType) && $this->FType == 'password' ) $valor = '******';

        if ($valorid != $valor) {
	    
            $tagParameter['valor']=htmlentities(utf8_decode($valorid));
	    if ($tagParameter['valor'] == '')
            $tagParameter['valor']=htmlentities($valorid);
        }

        // NO ZERO
        if ( isset($this->noZero) && $this->noZero == 'true' &&
            (   $valor  === 0  || $valor === '0' || $valorid==0)) {
            $valor = '';
        }

        if (!$limitSize) {
            $force = (isset($uiClass->Datos->forzado))?$uiClass->Datos->forzado:'';
            if (isset($uiClass->disabledCellId) && $uiClass->disabledCellId == true &&
                isset($uiClass->Datos->detalle) &&
                $detallecampo == '') {

                unset($tagParameter['valor']);

            }

            if (isset($uiClass->disabledCellId) && $uiClass->disabledCellId == true &&
               !isset($uiClass->Datos->detalle) &&
               $force != true)
                unset($tagParameter['valor']);

        }

        //////////////////////////////////////
        //////////////  STYLES  //////////////

        ////////////////////////////////////
        // Si el Campo representa un color
        ////////////////////////////////////
        $style2 ='';
        if (isset($this->color) && $this->color == 'true') {
            $style2 ='background-color:'.$valor.'; ';
            if ($this->transparent == 'true') {
                $style2.='color:'.$valor.'; ';
            }
        }

        // custom Font
        if (isset($this->font) && $this->font == 'true') {
            $style2.='font-family:'.$valor.'; ';
        } else {
            if (isset($this->font) && $this->font != 'true' && $this->font != '') {

                $rowtemp = $uiClass->Datos->TablaTemporal->getRowData($orden);
                if (isset($rowtemp[$this->font])) {
                    $valorFont=$rowtemp[$this->font];
                    $style2.='font-family:'.$valorFont.'; ';
                }
            }
        }
        ////////////////////
        // apply styles
        ///////////////////
        if (isset($this->noshow) && $this->noshow == 'true')
            $this->style = 'display:none;';

        $comboStyle  = (isset($this->style)) ?  $this->style :'';
        $comboStyle .= (isset($this->colstyle)) ?  $this->colstyle :'';
        $comboStyle .= (isset($style2)) ?  $style2 :'';

        if ($comboStyle != '') {
            $tagParameter['style'] = $comboStyle;
        }

        if (isset($this->contExterno->xml)) {
            $tagParameter['obj']=$this->contExterno->xml;
        }

        if (isset($this->PDFanchofijo)) {
            $tagParameter['PDFanchofijo']= $this->PDFanchofijo;
        }

        if (isset($uiClass->Datos->sumaCampo[$this->NombreCampo])) {
            $tagParameter['align'] = 'right';
        }

        if (isset($this->tdWidth))
            $tagParameter['width'] = $this->tdWidth;

        if (isset($this->tooltip))
            $tagParameter['title'] = $this->tooltip;

        //////////////////////////////////////
        // Valor por defecto autoactualizable
        /////////////////////////////////////

        if (isset($this->valauto) && $this->valauto !='') {
            $tagParameter['valauto'] = htmlspecialchars($this->valor);
        }

        if (isset($uiClass->hasFieldNameReference) && $uiClass->hasFieldNameReference == true || $fieldNameReference == true) {

            $tagParameter['campo']=$nomcampo;
        }
        //////////////////////
        // Add classes to tag
        /////////////////////
        if ( (isset($this->esClave) && $this->esClave)) {
            $tagParameter['class'] ='esclave ';
        }

        if (isset($this->liveSum))
            $tagParameter['class'] .= 'liveSum ';

        if (isset($this->Class))
            $tagParameter['class'] .= $this->Class.' ';

        /////////////////////
        // build Table Data
        /////////////////////
        $tableData = '';
        if ( (isset($this->noshow) && $this->noshow == 'true' ) || $uiClass->norepeat == true) {

            $tagParameter['valor']=htmlspecialchars($valorid);

                if ($tag != '') {
                    $tagParameter['border'] = '0';
                    $tagParameter['size']   = '0';

                    $tableData .= Html::Tag( $tag, '', $tagParameter);
                }
        } else {

            if ($imgIcon !='') {
            // Valor de la tabla
                $tagParameter['valor']=htmlspecialchars($valorid);

                if ($tag != '') {

                    $tableData .= Html::Tag( $tag, $imgIcon, $tagParameter);
                }

            } else {

                $searchString = (isset($uiClass->Datos->lastSearchString))? $uiClass->Datos->lastSearchString:'';

                if ($searchString != '' && $this->TipoDato != 'file')
                    $valor = resaltarStr($searchString , $valor);

                if ($tag != '') {
                    $tableData .= Html::Tag( $tag, $valor, $tagParameter);
                } else {
                    $tableData .= $valor;
                }

            }
        }

        return $tableData;
    }

    /**
      Obtiene la cadena a insertar en una sentencia SQL del campo
     */
    public function getSelectSQL($tabla = '')
    {
    // hack para nombres de campos vacios
        if ($this->NombreCampo =='')    return '';

        if (isset($this->TablaPadre))	$tabla = $this->TablaPadre;
        if (isset($this->alias) && $this->alias != '')	$tabla = $this->alias;

        if ($tabla != '')		$tabla .= '.';

        // Si es un valor local
        $comillas = '';
        if (isset($this->local)) {
            $tipo = Types::getTypeXSD($this->TipoDato, 'xsd:string');
            if ($tipo == 'xsd:string') {
                if (isset($this->valor) && substr($this->valor, 0, 1) != '"' && substr($this->valor, 0, 1) != "'")
                    $comillas= '"';

            }
            if (isset($this->valor) && $this->valor != '') {
                $sql = '('.$comillas.$this->valor.$comillas.')'.' ';

                if ($this->TipoDato == 'date') {
                    /* invierto fechas */
                    $val = $this->getValorGen($this->valor, $this->TipoDato);
                    $sql = '(\''.$val.'\')'.' ';
                }
            } else
                $sql = '('."''".')'.' ';
        } else
            $sql = $tabla.$this->NombreCampo.' ';

        if (isset($this->local) && $this->local =="true") $sql .= ' as ';

        // Si es una expresion SQL le agrega parentesis
        if (isset($this->parentesis) && $this->parentesis == 'false') {
            $par1 = '';
            $par2 = '';

        } else {
            $par1 = '(';
            $par2 = ')';
        }
        if (isset($this->Expresion))	$sql = $par1.$this->Expresion.$par2.' as ';
        if (isset($this->expresionSql) && $this->expresionSql == 'true')	$sql = '("") as ';

        // Si el campo tiene etiqueta se agrega un Alias en el SQL
        if (isset($this->alias) && $this->alias != '') {
            $sql .= $this->alias.' ';
        } else {
            /*
            if ($this->Etiqueta != '')
                $sql .= '"'.$this->Etiqueta.'"'.' ';
            else
            */
                $sql .= '"'.$this->NombreCampo.'"'.' ';
        }

        return $sql;
    }

    /**
      Obtiene Las condiciones inherentes al campo para aplicar en una sentencia SQL
     */
    public function getCondiciones($tabla = '')
    {
        $wheres = '';
        if (isset($this->local)) return;
        if ($tabla != '') $tabla .= '.';

        $campo = $this->NombreCampo;

        if (isset($this->Expresion)) {

            //$campo = $this->Etiqueta;
            if ($this->having == 'false') {
                $campo = $this->Expresion;
            } else
                if ($campo == '') $campo = $this->Expresion;
	    
            $tabla = '';
        }

        // Armo condiciones agrupadas

        if (isset ($this->Condiciones)) {

            foreach ($this->Condiciones as $key => $condicion) {
                $grupo = ($condicion->grupo != '')? $condicion->grupo : 0;
                $condiciones[$grupo][]= $condicion;

            }
            $cant = count($condiciones);
            $numberOfConditions = 0;
            if ($condiciones)
                foreach ($condiciones as  $grupoCondiciones) {
                    $numberOfConditions++;
                    $first = true;
                    $stringWhere = '';
                    foreach ($grupoCondiciones as $key => $valor) {
                        $cadena = $valor->armaCondicion($tabla.$campo, $first);
                        $stringWhere .= $cadena;
                        if ($stringWhere != '')
                            $first = false;
                    }
                    if ($stringWhere != '') {
                        $wheres .= ' ('. $stringWhere.') ';

                    }
                    if ($numberOfConditions < $cant) $wheres .= ' or '; // ver como parametrizar este operador
                }

        }

        return $wheres;
    }

    /**
      Oculta/Muestra el campo/columna en la consulta
     */
    public function setOculto($val)
    {
        if ($val != '')
            $this->Oculto = $val;
    }
    /**
      Devuelve si es un campo Oculto
     */
    public function esOculto()
    {
        if (isset($this->Oculto))
            return $this->Oculto;
        else return false;

    }

    public function setArbol($val)
    {
        $this->Arbol = $val;
    }

    /**
    vincula el detalle del campo con el valor del mismo para pasar como parametro
    */
    public function getUrlVariableString($valor, $encode=true)
    {
        $str = '';
        $concat = '';
        if (isset($this->Detalle) && $this->Detalle != '') {

            foreach ($this->Detalle as $deta) {
                  if (!$encode)
                    $str .= $concat.$deta.'='.$valor;
                  else
                    $str .= $concat.$deta.'='.urlencode($valor).'';
                $concat = '&';
            }
        }

        return $str;
    }
    /**
      Devuelve los posibles operadores lógicos para los diferentes tipos de datos
     */
    public function getOperadores()
    {
        $operadores = array ('>' => 'Mayor que ', '<' => 'Menor que ', '=' => 'igual que ', '!=' => 'Distinto de ');

        return $operadores;
    }

    /**
     * Update Data conditions for que Query of the Inner Datacontainer
     */
    public function refreshInnerDataContainer($dataContainer, $row = null)
    {
    // Obtengo los Datos de los parametros definidos para el Contenedor Externo Embebido
        if ($this->paring !='') {
            foreach ($this->paring as $destinodelValor => $origendelValor) {
                $comillas = '';

                $fieldSource = $dataContainer->getCampo($origendelValor['valor']);
                if ($fieldSource)
                    $valorDelCampo = $fieldSource->getValor();

                if ($row != null) {
                    $valorDelCampo = $row[$origendelValor['valor']];
                    if ($valorDelCampo == '') $valorDelCampo = '0';

                }

                if (isset($valorDelCampo) ) {
                //if (isset($valorDelCampo) && $valorDelCampo != '') {

                    if ($fieldSource->TipoDato == 'varchar') $comillas = '"';
                    $operador = ($origendelValor['operador'] != '')?$origendelValor['operador']:'=';

                    $reemplazo = $origendelValor['reemplazo'];
                    if ($reemplazo != 'false')
                        $reemplazo ='reemplazo';
                    else $reemplazo = '';

                    $this->contExterno->addCondicion($destinodelValor, $operador, $comillas.$valorDelCampo.$comillas, 'and', $reemplazo, false);
                    $this->contExterno->setFieldValue($destinodelValor, $valorDelCampo, 'both');
                    $field = $this->contExterno->getCampo($destinodelValor);
                    if ($field)
                        $field->setValorOriginal($valorDelCampo);

//                    $this->contExterno->setCampo($destinodelValor, $valorDelCampo);
//                    $this->contExterno->setNuevoValorCampo($destinodelValor, $valorDelCampo);
                }
            }
        }
    }

    /**
     * input attributes
       
     * @return array input attributes
     */
    private function inputAttributes(){

        if (isset($this->validar) && $this->validar == 'false')     
            $arrayAttributes['novalidar']    = "true";                  // force NO validation (Tested)

        if (isset($this->preparamCampo ) && $this->preparamCampo != '' ){ 
            $arrayAttributes['oblig']    = "true";                      // (Tested)
	}
        
        if (isset($this->cuit) && $this->cuit == 'true')        
            $arrayAttributes['cuit']     = "true";                      // force cuit check (Tested)
        
        if (isset($this->onformchange) && $this->onformchange != '')    
            $arrayAttributes['onformchange'] = $this->onformchange;     // on form change event (implemented?)
        
        if (isset($this->valauto) && $this->valauto !='')       
            $arrayAttributes['valauto']  = $this->valauto;              // Automatic value field (tested)
        
        if (isset($this->required)){

	    if ( $this->required == 'true'){    
        	$arrayAttributes['required']     = "required";               // Required field (tested)
        	$arrayAttributes['oblig']    = "true";                      // (Tested)
	    } else {
		unset($arrayAttributes['required']);
		unset($arrayAttributes['oblig']);
	    }
	}      

        if (isset($this->min) && $this->min !='')           
            $arrayAttributes['min']      = $this->min;           // min Value (tested)
        
        if (isset($this->max) && $this->max !='')           
            $arrayAttributes['max']      = $this->max;           // max Value (tested)
        
        if (isset($this->idlbl) && $this->idlbl !='')           
            $arrayAttributes['idlbl']    = $this->idlbl;         // Label id (tested)

        


        if (isset($this->esClave) && $this->esClave) {
            $arrayAttributes['esClave'] = "true";    // force required for key index fields (tested)
            $arrayAttributes['required'] = "required";
        }

        if (isset($this->pintado) && $this->pintado =='true')       
            $arrayAttributes['pintado']  = "true";                           // ?


        if (isset($this->required) && $this->required =='false')    
            unset($arrayAttributes['required']);                             // force unset required (tested)

     
        if (isset($this->preventLoop)) {
            $arrayAttributes['preventLoop'] = 'true';  // (tested)
        }

        $arrayAttributes['align'] = 'left';


        // Error Message
        if (isset($this->errorMessage)) {
            $arrayAttributes['errorMessage']= $this->errorMessage;  // (tested)
        }

        // Defaul Value ghost behavior
        if (isset($this->defaultValue)) {
            $arrayAttributes['defaultValue']= $this->defaultValue;  // (tested)
        }

        if (isset($this->mask)) {
                $arrayAttributes['mask']    = $this->mask;  // (tested)
        }

        $arrayAttributes['type'] = 'text';                   // default type

        return $arrayAttributes;
    }

    /**
     * Renders Input field
     
     * @param  object $uiClass      User Interfase Class
     * @param  string $formName     Form Name
     * @param  string $prefijoId    [description]
     * @param  string $valor        field Value
     * @param  string $modoAbm      [description]
     * @param  string $idContenedor [description]
     * @param  string $atrib        [description]
     * 
     * @return string field Input
     */
    public function renderInput($uiClass, $formName, $prefijoId = '', $valor = '', $modoAbm = '',  $idContenedor = '', $atrib='')
    {

        if (isset($this->conLabel) && $this->conLabel =='true') {
            $this->idlbl=UID::getUID('',true);
           // $id2 = 'id="'.$this->idlbl.'"';
        }

        $fieldName = $this->NombreCampo;
        $size      = $this->Size;

        if (is_array($atrib)) $arrayAttributes = $atrib;

        $arrayAttributes['form'] = $uiClass->Datos->xml;                // Name of the form
        $arrayAttributes['name'] = ($prefijoId == '')?$fieldName:$prefijoId;   // input name

        /////////////////////////
        // get input field attributes
        ////////////////////////
        $arrayAttributes = array_merge($arrayAttributes, $this->inputAttributes());

        if (isset($this->autoing) && $this->autoing =='true') {
            $arrayAttributes['autoing'] = "true";                       // autoload in forms
            $uiClass->Datos->autoing="true";
        }


        $valop = (isset($this->valop))?$this->valop:'false';

        if ($valop != 'true') 
            $opciones = (isset($this->opcion))?$this->opcion:null;



        if (get_class($this) == 'Field')
            $ayuda = htmlentities(ucfirst($this->getAyuda()), ENT_QUOTES, 'UTF-8');

        $esClave = false;
        if (isset($this->esClave) && $this->esClave == 'true') {
            $esClave = true; // La clave de busqueda habilita al campo a buscar los valores de los demas con un SQL
            //$ayuda .= ' (* '.$uiClass->i18n['required'].')';
        }
        $esClaveBusqueda = $esClave;

        //////////////////
        // disable field
        //////////////////
        $deshabilitado = '';
        if (isset($this->deshabilitado))
            $deshabilitado = $this->deshabilitado;

        if (isset($this->habilita) && $this->habilita != '') {
            $lastValue = $uiClass->Datos->getCampo($this->habilita)->ultimo;

            if ($lastValue == 0 ||
                    $lastValue == '0' ||
                    $lastValue === 'false' ||
                    $lastValue === false) {

                $deshabilitado='true';
            }
        }

        if (isset($this->conditionalDisplay)) {
            $lastValue = $uiClass->Datos->getCampo($this->conditionalDisplay)->ultimo;
            if ($lastValue == 0 ||
                    $lastValue == '0' ||
                    $lastValue === 'false' ||
                    $lastValue === false) {
                return false;
            }
        }

        $tammax = (isset($this->Tammax))?$this->Tammax:null; // tamaño manimo

        //////////////////////////////////
        // ASSIGN ID
        // REFACTOR TO SEPARATE CODE
        //////////////////////////////////
        // note: 
        $id=($prefijoId == '')? $this->uid:$prefijoId;
       

        $uniqid2 = UID::getUID($id, true);

        $this->uid2 = $uniqid2;
        // for date filters
        if ($this->TipoDato != 'date') {
            $this->uid2 = $uniqid2;
        } else {
            $this->uid2 = (isset($this->uid2))? $this->uid2:$uniqid2;
            $uniqid2 = $this->uid2;

        }

        //$arrayAttributes['id']         = $id;
        $arrayAttributes['id']       = $uniqid2;

        if ($uiClass->Datos->tipoAbm != 'grid')
            $arrayAttributes['uid']      = $uniqid2;

        $arrayAttributes['title']    = isset($ayuda)?$ayuda:'';

        if ($deshabilitado != 'true' && isset($ayuda))
            $arrayAttributes['placeholder']  = $ayuda;



        // acciones para los eventos
        $blur   = '';
        $change = '';

        $keyup  = '';

        $actualizar2 = '';
        $javascripAejecutar= '';
        $actualizarSelect2 = '';
        $actualizarSelect = '';
        $calculojsBLUR = '';
        $calculojsKEY  = '';


        //Calculo en Javascript
        $calculojs = '';

        if (isset($this->jseval) && $this->jseval != '') {
            //    $jscabecera = ", ''";
            $jscabecera = "";
            if (isset ($uiClass->Datos->CabeceraMov)) {
                foreach ($uiClass->Datos->CabeceraMov as $cabecera) {
                    $jscabecera = "Form".$cabecera->xml;
                }
            } else {
                if (isset($uiClass->Datos->xmlReferente)) {
                    $jscabecera = "Form".$uiClass->Datos->xmlReferente;
                }
            }

            if ($jscabecera != '') {
                $formaCalc = $jscabecera;
            } else {
                $formaCalc = $uiClass->Datos->xmlOrig;
            }

            if ($uiClass->Datos->tipoAbm == 'grid') {
                $formName = 'Form'.$uiClass->Datos->xmlpadre;
            }

            foreach ($this->jseval as $campodestino => $jseval) {

                //  Modificar esto, esta mal implementado ok je
                $arrayJseval[]      = $jseval;
                $arrayJsevaldest[]  = $campodestino;
                if (isset($this->jsevalactxml[$campodestino]) && $this->jsevalactxml[$campodestino]!='false')
                    $formaCalc = $this->jsevalactxml[$campodestino];
                $actxml= (isset($this->jsevalactxml[$campodestino]) && $this->jsevalactxml[$campodestino]=='false')?'false':'true';

                $calculojs  = " calculojs('".$jseval."', '".$campodestino."', '".$formName."', $('#".$uniqid2."')[0] , '"            . $formaCalc. "' ,  '".$actxml."' ); ";
                //$calculojs2 = " calculojs('".$jseval."', '".$campodestino."', '".$formName."' $('#".$id."')[0] , '" . $formaCalc. "' ,  '".$actxml."' ); ";
                //            formula, destino, formulario, obj, formExtra, act, actxml11

                if ($campodestino != '__EVAL') // no valido todo el tiempo
                    $calculojsKEY .= $calculojs;

                $calculojsBLUR .= $calculojs;

            }
            $arrayAttributes['jseval']     = "new Array('" . implode("','", $arrayJseval     ) . "')";
      //      $arrayAttributes['jsevaldest'] = "new Array('" . implode("','", $arrayJsevaldest ) . "')";

            $arrayAttributes['jsevaldest'] = htmlspecialchars(json_encode($arrayJsevaldest));


            $keyup .= $calculojsKEY;
            $blur  .= $calculojsBLUR;
            $change.= $calculojsBLUR;
        }

        // Añado el javascript necesario para extraer substrings del input y
        // asignarlo a varios campos

        if (isset($this->jsextract) && $this->jsextract != '') {
            foreach ($this->jsextract as $destino => $posiciones) {
                $arrayDestinos[] = $destino;
                $arrayposini[]   = $posiciones['posini'];
                $arrayposfin[]   = $posiciones['posfin'];
            }
            $posicionesIni  = "new Array('" . implode("','", $arrayposini   ) . "')";
            $posicionesFin  = "new Array('" . implode("','", $arrayposfin   ) . "')";
            $arrayDestino   = "new Array('" . implode("','", $arrayDestinos ) . "')";

            $jsextract = 'jsextract(this, '.$arrayDestino.', '.$posicionesIni.' ,'.$posicionesFin.');';
            $change .= $jsextract;
        }

	$focusSelect= '';

        if (isset($this->jsExec))
            $focusSelect= $keyup;

        if ($uiClass->Datos->tipoAbm == 'grid' && $this->suma =="true" || (isset($this->calculatesum) && $this->calculatesum == 'true')) {
            $calculojs      = 'Histrix.calculoTotal(this);';
            $calculojsBlur  = 'Histrix.calculoTotal(this, null, true);';

            //    $javascripAejecutar = 'Histrix.calculoTotal(undefined, \''.$uniqid.'\' );';
            $keyup  .= $calculojs;
            $change .= $calculojs;
            $blur   .= $calculojsBlur;
        }

        if (isset($this->activador) && $this->activador !='') {
            if ($uiClass->Datos->getCampo($this->activador)->valor != 1) {
                $deshabilitado='true';
            }

        }
        if ($uiClass->Datos->tipoAbm == 'grid') {
            $style.='width:100%;';
            $style.=$this->colstyle;
        }
        // Si el Campo representa un color
        if (isset($this->color) && $this->color == 'true') {
            $arrayAttributes['escolor'] = "true";
            $style.='background-color:'.$valor.'; ';
            if (isset($this->transparent) && $this->transparent == 'true') {
                $style.='color:'.$valor.'; ';
            }
        }

        // copio el valor del campo a otro en via javascript
        if (isset($this->copia) && $this->copia !='') {
//      loger($uiClass->Datos->getCampo($this->copia));
            //$copyId= $uiClass->Datos->getCampo($this->copia)->uid;
            $change.=" copiavalorcampo('".$this->copia."', this ,'".$uiClass->Datos->xml."' ); ";
            $blur .= $change;
        }
        
        // deshabilito el campo para updatear si es una expresion
        if (isset($this->Expresion) && $this->Expresion != '' || $deshabilitado == 'true') {

            if (isset($this->enablecopy) && $this->enablecopy == "true") {
		// enable 
            }
	    else {
                //   if we use disabled, we cant copy
                $arrayAttributes['disabled']='disabled';

	    }
            
            $arrayAttributes['readonly']='readonly';
        }
        // a menos que lo quiera explicitamente habilitado por alguna extra�a razon
        if ($deshabilitado == 'false') {

            unset($arrayAttributes['readonly']);
            unset($arrayAttributes['disabled']);
        }
        // Si la ficha esta en modo lectura o para los campos NO clave de los ABM
        if ($modoAbm == 'readonly') {

            $arrayAttributes['readonly']='readonly';
            $arrayAttributes['class']="readonly";
        }

        if (isset($this->addClass) && $this->addClass != '') {
            $arrayAttributes['class'] .= $this->addClass;
        }

        $busca = false;
        if (get_class($this) == 'Field') {
            if (($this->getContenedorAyuda()) &&
             $this->searchRecord != 'false' &&
             ( !isset($this->autocompletar) || isset($this->autocompletar) && $this->autocompletar != 'true')) {

                $busca = true;

                $optionsArray['instance']   = $uiClass->Datos->getInstance();

                //$postOptions = Html::javascriptObject($optionsArray);
                
                $jsonOptions = htmlspecialchars(json_encode($optionsArray));
                
                $change .= "buscaregistro ( '".'desc_'.$id."', '".$fieldName."' , this, '".$uiClass->Datos->xml."', null, null, $jsonOptions ); ";
                $blur  .= $change;
            }
	}

        if (isset($this->Parametro['incdec']) && $this->Parametro['incdec']) {
            $arrayAttributes['class'] = 'SpinButton';
        }
        // Linkexterno
        if (isset($this->linkExt) &&  $this->linkExt != '') {
            $linkExt= '<span onClick="Histrix.loadInnerXML(\''.$uiClass->Datos->xml.'\', \''.$this->linkExt.'\' ,null, \''.$this->linkExtTit.'\', \''.$uiClass->Datos->xml.'\', \''.$id.'\');"><img  title="Busqueda Externa" src="../img/applications-internet.png"></span>';
        }


        $icoFile ='';
        $salida = '';
        $browsePref = '';
        $browse = '';
        $linkExt = '';


        //////////////////////////////////////////////////////////////
        ///  get Attributes and custom values form FieldType Classes
        //////////////////////////////////////////////////////////////

        $fieldTypeClass = 'FieldType_'.$this->TipoDato;
        $customDisplay = false;
        //if (is_file(dirname(__FILE__).'/FieldType/'.$fieldTypeClass.'.php')) {
        if ($this->TipoDato != ''){
            if (class_exists($fieldTypeClass)) {
                //  $Type = new $fieldTypeClass($this);

                $arrayAttributes['align']   = constant($fieldTypeClass.'::ALIGN');
                $arrayAttributes['dir']     = constant($fieldTypeClass.'::DIR');

                if (defined($fieldTypeClass.'::INPUT')) {
                    $arrayAttributes['inputType'] = constant($fieldTypeClass.'::INPUT');
                }

                // TODO CHECK IMPLEMETATION maybe init methos is not needed
                if (defined($fieldTypeClass.'::CUSTOM')) {
                    //$objectFieldType = new $fieldTypeClass($this);
                    //$objectFieldType::init($this);
                    //
                    $fieldTypeClass::init($this);
                    $customDisplay = true;
                }

                // get input attributes from fieldType Class
                if (method_exists($fieldTypeClass ,'inputAttributes')) {

                    $fieldTypeAttributes = $fieldTypeClass::inputAttributes($valor, $this);

                    if (is_array($fieldTypeAttributes)){
                        $arrayAttributes = array_merge($arrayAttributes, $fieldTypeAttributes);
                    }
                }

                // extra Info to be implemented
                //  $dat =  $fieldTypeClass::extraData();

            } else {

                loger('falta: '.dirname(__FILE__).'/FieldType/'.$fieldTypeClass.'.php', 'fieldtypes.log');
            }
        }
        if (isset($this->FType)   && $this->FType   =='password')           // TODO: REMOVE Type="password" support
            $arrayAttributes['type']     = 'password';                       // type password
 
        switch ($this->TipoDato) {

            case "date" :

                if (( ($arrayAttributes['readonly'] != '') && ($arrayAttributes['disabled'])   || $deshabilitado != 'true') && $arrayAttributes['hidden'] != 'true' ) {
                    $keyup .= ' this.value=formateafecha(this.value); ';
                    $blur .= ' isDate(this); ';

                }
                $size =10;
            break;
	    case "check":
		$blur = '';
	    break;

            case "dir":

                if ($this->FType != '' ) $tipoFile = $this->FType;
                if ($this->url != '' )   $url = $this->url;

                $arrayAttributes['type'] = 'hidden';

                //if ($this->browseButton != 'false') {

//                if ($uiClass->Datos->path != '' || $this->path != '') {

                    // Path for uploading files
                 if ($uiClass->Datos->path != ''  )
                    $url .= $uiClass->Datos->path.'/';

                    // Create choice button
                    $choose = 'Elegir';
                    $btnBrowse = new Html_button('', '../img/imgfolder.gif' ,$choose );
                    $btnBrowse->addParameter('title', $choose);

                    if ($this->browse != '') {
                        $btnBrowse = new Html_button("Archivo", null ,"Archivo" );
                        $btnBrowse->addParameter('pathObj','');
                    }

                    // Add Object Path
                    if ($this->path !='') {
                        $pathField = $uiClass->Datos->getCampo($this->path);
                        if (is_object($pathField)) {
                            $objUrl = $pathField->valor;
                        } else {
                            $objUrl = $this->path;
                        }

                        if ($objUrl !='') {
                            $url .= $objUrl.'/';
                        } else {
                            $btnBrowse->addParameter('pathObj',$pathField->uid2);
                        }
                    }

                    $btnBrowse->addParameter('path',$url);
                    $btnBrowse->addEvent('onclick', 'BrowseServer2(\''.$uniqid2.'\', \''.$tipoFile .'\', \''.$url.'\',\''.$uniqid2.'\', this)');
            //    }

                    if ($this->deshabilitado != 'true') {
                        $browse =  $btnBrowse->show();
			
			if ($this->drop != 'false'){
                    	    $browse .= '<div class="dropfile" pathObj="'.$pathField->uid2.'" destinationDir="'.$url.'">'.$uiClass->i18n['dropfile'].'</div>';
			}

                    }

                break;
        }


        if ($esClaveBusqueda && $modoAbm != 'readonly' && 
                ( isset($this->deshabilitado) && $this->deshabilitado != 'true' || !isset($this->deshabilitado))
                && $uiClass->Datos->tipoAbm == 'ficha') {
            $arrayAttributes['class']="clave";
        }

        if ($uiClass->Datos->tipoAbm=='cabecera' && $busca != true) {
            $blur .= 'llenoCabecera(\''.$formName.'\', \''.$uiClass->Datos->xml.'\' , \''.$uiClass->Datos->xmlOrig.'\', this);';
        }

	$formOptions = '';
        // Actualización de Campos
        if (isset($this->actualizarCampo) && $this->actualizarCampo != '') {
            foreach ($this->actualizarCampo as $idCampo) {

                $arrayAttributes['class'] .= ' refreshable';
                $arrayDestino = $this->actualizarDestino[$idCampo];
                if ($arrayDestino != '') {
                    $comboDestino = $idCampo;
                    $camposDestino = "new Array('" . implode("','", $arrayDestino) . "')";

                    $xmlDestino =($this->actualizarCampoXml[$idCampo] != '')?$this->actualizarCampoXml[$idCampo]:$uiClass->Datos->xml;

                    if (isset($this->actualizarFilter[$idCampo]) && $this->actualizarFilter[$idCampo] == 'false') {
                        $formOptions='filter:false';
                    }
                    $actualizarSelect2 .= ' actualizarCombo2(this, \''.$comboDestino.'\' ,'.$camposDestino.' , \''.$xmlDestino.'\', \''.$formName.'\',{'.$formOptions.'});';

                } else {
                    // solamente se actualiza el valor del campo destino
                    $actualizarSelect.= 'setValorCampo(this.form,\''.$idCampo.'\' , this.value, \''.$uiClass->Datos->xml.'\'); ';
                }
            }
        }

        // Detecto que tipo de Input corresponde mostrar
        $inputType = 'textBox'; // Por defecto

        if ($size >= 100 && $uiClass->tipo != 'ing' && $browse == '')  
            $inputType = 'textArea';

        //////////////////////////////////////
        // Get Input Type from Class
        //////////////////////////////////////
        if (isset( $arrayAttributes['inputType'])) {
            $inputType = $arrayAttributes['inputType'];

        }

        //////////////////////////////////////
        // force Input Type from field attribute
        //////////////////////////////////////
        if (isset( $this->inputType)) {
            $inputType = $this->inputType;

        }


        //////////////////
        // force Select
        //////////////////
        if (count($opciones) > 0 && $this->valop != 'true' && $arrayAttributes['inputType']!='radio'){     
            $inputType = 'select';
        }

        // unset var not to be used later
        unset($arrayAttributes['inputType']);


        if ( ( isset($this->hidden) && $this->hidden == 'true') || 
            (isset($this->noshow) && $this->noshow == 'true' && $modoAbm != 'filtro')) {
            $arrayAttributes['type'] ='hidden';
        }

        //////////////////////
        //  js refresh
        //////////////////////

        $refrescaOrig= '';
        if ( isset($uiClass->Datos->xmlOrig) && $uiClass->Datos->xmlOrig != '' && isset($this->refresh) && $this->refresh=="true") {

            $refrescaOrig = "llenoCabecera('Form".$uiClass->Datos->xml."', '".$uiClass->Datos->xml."' , '".$uiClass->Datos->xmlOrig."', this); ";
            $refrescaOrig .= 'grabaABM(\'Form'.$uiClass->Datos->xmlOrig.'\', \'update\',\''.$uiClass->Datos->xmlOrig.'\' , \''.$uiClass->Datos->xmlOrig.'\'); ';
            $change .= $refrescaOrig;
        }

        // refresh Attribute
        if (isset($uiClass->Datos->esInterno)) {
            $arrayAttributes['innerform']= 'Form'.$uiClass->Datos->idxml;
        }

        // Refresh Attributes for input fields
        if (isset($this->refresh)) {

            $arrayAttributes['refresh']= $this->refresh;
            
            $orden = $uiClass->_rowId;

            $change = 'setCampoTabla('.$orden.' , \''.$this->NombreCampo.'\', this , \''.$uiClass->Datos->xml.'\', '.$arrayAttributes['refresh'].', \''.$uiClass->Datos->xmlOrig.'\',  \''.$uiClass->Datos->getInstance().'\');';

        }

        // Set Tabindex
        if ($deshabilitado != 'true') {
            $tabindex = (isset($this->lastTabindex) && $this->lastTabindex != ''
                            && $uiClass->Datos->tipoAbm != 'grid' && $uiClass->Datos->tipoAbm != 'liveGrid'
                    )?$this->lastTabindex:$uiClass->tabindex();
        } else {
            $tabindex='';
        }

        if (isset($atrib['editrow']) && $atrib['editrow'] ==true) {
            if (strlen($valor) > 0)
                $size = strlen($valor) + 1;
        }


        // input Type
        // Muestro segun el Tipo de input
        switch ($inputType) {
            case 'button':
                if ($valor != '') {
                    $inputBox = new Html_button(Types::removeQuotes($valor), '' ,$valor );

                    $styles  = (isset($this->style))?$this->style:'';
                    $styles .= (isset($this->Formstyle))?$this->Formstyle:'';

                    $inputBox->addParameter('style', $styles);
                    $inputBox->addParameter('class', $this->Class);

                    unset($styles);
                    // add custom Javascript Events
                    if ($this->jsfunction)
                        foreach ($this->jsfunction as $jsevent => $jsfunctions) {
                            foreach($jsfunctions as $nfunc => $jsfunction)
                                $inputBox->addEvent($jsevent, $jsfunction, true); // append function

                        }

                    $salida .= $inputBox->show();
                }

            break;
            case 'radio':
            case 'file':
            case 'geoPoly':
            case 'geoPoint':
            case 'editor':
            case 'simpleditor':

                if ($this->customRowName === true) {
                    $arrayAttributes['name'] .= $uiClass->_rowId;
                }

                $arrayAttributes['onchange'] .= $actualizarSelect2;
                $arrayAttributes['size']      = $size;
                $arrayAttributes['form']      = $formName;            
                $arrayAttributes['tabindex']  = $tabindex;
                 
                $salida .= $fieldTypeClass::renderInput($valor, $this, $arrayAttributes, $uiClass, $opciones);                

            break;

            case 'textArea':

                $inputBox = new Html_textArea();
                $inputBox->value    = $valor;
                $inputBox->size     = $size;
                if ($this->cols != '')
                    $inputBox->cols     = $this->cols;

                $inputBox->tabindex = $tabindex;

                $inputBox->Parameters = $arrayAttributes;
                $inputBox->addParameter('form', $formName);
                $inputBox->addParameter('maxlength', $this->maxlength);

                $inputBox->addParameter('style', $this->style);


                // add custom Javascript Events
                if ($this->jsfunction)
                    foreach ($this->jsfunction as $jsevent => $jsfunctions) {
                        foreach($jsfunctions as $nfunc => $jsfunction)
                            $inputBox->addEvent($jsevent, $jsfunction, true); // append function
                    }
                $salida .= $inputBox->show();

                break;
            case 'select':

                unset ($arrayAttributes['dir']);

                $inputBox = new Html_select($opciones);
                $inputBox->Parameters=$arrayAttributes;

                if ($this->selectExpand == 'true') {

                    $sizeExpand = 5;

                    $inputBox->addParameter('size', $sizeExpand);

                }
                if (isset($this->noupdate) && $this->noupdate == 'true') {
                    $inputBox->addParameter('noupdate', $this->noupdate);

                }
                if ($this->multiple=="true")
                $inputBox->addParameter('multiple', 'multiple');


                $inputBox->addParameter('style', $this->style);
                $inputBox->addParameter('form', $formName);

                $inputBox->addEvent(array('onblur','onchange'), $blur);

                $inputBox->addEvent('onchange', $change, true);

                $inputBox->addEvent('onchange', $actualizarSelect, true);
                $inputBox->addEvent('onchange', $actualizarSelect2, true);

                $inputBox->addEvent('onfocus', $focusSelect);

                // add custom Javascript Events
                if (isset($this->jsfunction))
                    foreach ($this->jsfunction as $jsevent => $jsfunctions) {
                        foreach($jsfunctions as $nfunc => $jsfunction)
                            $inputBox->addEvent($jsevent, $jsfunction, true); // append function
                    }

                $javascripAejecutar .= $focusSelect;

                $inputBox->value     = $valor;
                $inputBox->xml       = $uiClass->Datos->xml;
                $inputBox->xmlOrig   = $uiClass->Datos->xmlOrig;
                $inputBox->size      = $size;
                $inputBox->Formato   = $this->Formato;
                $inputBox->tabindex  = $tabindex;

                if ($deshabilitado == 'true') {
                    /* deshabilito los eventos? */
                    $inputBox->removeEvent(array('onfocus', 'onblur', 'onkeyup', 'onclick'));                    
                }

                $selectString = $inputBox->show();

                if ($inputBox->match != true) {
                    $this->valor = $inputBox->Campo_valor;
                    $this->nuevovalor = $inputBox->Campo_nuevovalor;
                }

                if (trim($actualizar2) != '') {
                    $selectString .= Html::scriptTag(array($actualizar2));
                }

                $salida .= $selectString;

                if ($this->contExterno->xml !='' && $this->contExterno->tipoAbm == 'abm-mini') {
                    $this->tooltip = $this->contExterno->titulo;
                    $formatoCampo = (isset($this->Formato))?$this->Formato:null;
                    $salida .= $uiClass->linkButton($this, '+', null, null, $formatoCampo);
                }
                break;


            case 'checkbox':
    $blur='';
                break;
            case 'number':
            case 'numeric':            
            case 'integer':
            case 'text':
            case 'textBox':

                $inputBox = new Html_textBox($valor, $this->TipoDato);
                if (isset($this->Parametro['incdec']) && $this->Parametro['incdec']) {
                    $arrayAttributes['class'] = 'SpinButton';
                }
                $inputBox->Parameters=$arrayAttributes;

                $styles  = (isset($this->style))?$this->style:'';
                $styles .= (isset($this->Formstyle))?$this->Formstyle:'';

                if ($styles != '')
                    $inputBox->addParameter('style', $styles);

                $inputBox->addParameter('form', $formName);
                $inputBox->addEvent('onchange', $change);
                $inputBox->addEvent('onblur', $blur);

                if ($deshabilitado == 'true')
                    $inputBox->addEvent('onchange', $blur);

                $inputBox->addEvent('onkeyup', $keyup);

                $inputBox->addEvent('onchange', $actualizarSelect2, true);

                if (isset($focusSelect))
                    $inputBox->addEvent('onfocus', $focusSelect);

                // add custom Javascript Events
                if (isset($this->jsfunction))
                    foreach ($this->jsfunction as $jsevent => $jsfunctions) {
                        foreach($jsfunctions as $nfunc => $jsfunction)
                            $inputBox->addEvent($jsevent, $jsfunction.'; ', true); // append function
                    }

                // NO SE EJECUTA SOLO...?
                //// VER SI EJECUTO ESTO SE PONE RE LENTO DONDE HAY ADE UNA UN MONTON DE CAMPOS!!!
                if (isset($this->jsExec) && $this->jsExec == 'true')
                    $javascripAejecutar .=$focusSelect;

                //$javascripAejecutar .= $uiClass->Datos->jQueryCalculationString($this);
                // echo $javascripAejecutar;

                $inputBox->value    = $valor;
                if (isset($this->valor))
                    $inputBox->Campo_valor  = $this->valor; // para las fechas

                $inputBox->valorCampo   = (isset($this->valorCampo))?$this->valorCampo:null; // No se para que VERRRRRR
                $inputBox->xml      = $uiClass->Datos->xml;
                if (isset($uiClass->Datos->xmlOrig))
                    $inputBox->xmlOrig  = $uiClass->Datos->xmlOrig;
                $inputBox->deshabilitado    = $deshabilitado;
                $inputBox->size     = $size;
                if (isset($this->forceSize) && $this->forceSize != '')
                    $inputBox->forceSize = true;

                if (isset($this->maxsize))
                    $inputBox->maxsize  = $this->maxsize;
                $inputBox->tammax   = $tammax;

                if (isset($this->Formato))
                    $inputBox->Formato  = $this->Formato;
                $inputBox->tabindex     = $tabindex;
                $inputBox->tipoAbm      = $uiClass->tipo;

		        if (isset($this->maxlength))
                    $inputBox->addParameter('maxlength', $this->maxlength);


                if (isset ($this->contAyuda)) {
                    $objAyudaXml = 'HLP_'.$fieldName;
                    //Histrix_XmlReader::serializeContainer($this->contAyuda, null, '_'.$objAyudaXml);
                    Histrix_XmlReader::serializeContainer($this->contAyuda);
                    
                    $div = '';
                    if (isset($this->autocompletar) && $this->autocompletar == 'true') {
                        $autominchars = ($this->autominchars != '')?$this->autominchars:3;
                        $max = ($this->contAyuda->limit!='')?$this->contAyuda->limit:0;
                        $div = '<div class="autocompletar" id="ayuda_'.$id.'" ></div>';
                        $phpDestino = "'process.php?xmldatos=".$objAyudaXml."&accion=help&autocomplete=true&idinput=".$fieldName.'&xmlOrig='.$uiClass->Datos->xml.'&instance='.$this->contAyuda->getInstance()."'";
                        $rowid = (isset($uiClass->_rowId))?$uiClass->_rowId:'null';
                        $javascripAejecutar = "$('#".$uniqid2."').autocomplete( ".$phpDestino.", {minChars:".$autominchars.", max:". $max ."}).result( function( event, data){ autoComplete(event, data, '".$uiClass->Datos->idxml."', $rowid);  }); ";
                        if (!isset($uiClass->_rowId))
                            $inputBox->sufijo .= '<img src="../img/remove2.png" onClick="$(\'#'.$uniqid2.'\').clear();"/>';

                    }
                    $inputBox->prefijo .=  $div;
                }

                if (get_class($this) == 'Field')
                    $thisContenedorAyuda = $this->getContenedorAyuda();

                if (isset($thisContenedorAyuda) && $thisContenedorAyuda != false  && $this->autocompletar != 'true') {
                    // Ver si se identifican bien las ayudas que no poseen xml asociado
                    $xmlHLP = $thisContenedorAyuda->xml;
                    Histrix_XmlReader::serializeContainer($thisContenedorAyuda);
                    
                    $xmlOrig = $uiClass->Datos->xml;
                    $xmlForm = $uiClass->Datos->xml;
                    if ($uiClass->Datos->xmlOrig != '') $xmlOrig = $uiClass->Datos->xmlOrig;

                    if ($uiClass->Datos->xmlpadre != '') {
                        $xmlOrig = $uiClass->Datos->xmlpadre;
                        $xmlForm = $uiClass->Datos->xmlpadre;

                    }
                    $styleAyuda = $thisContenedorAyuda->style;

                    $optionsArray['xmlOrig']    = $uiClass->Datos->xml;
                    $optionsArray['xmlHlp']     = $xmlHLP;
                    $optionsArray['helpStyle']  = $styleAyuda;
                    $optionsArray['xmlform']    = $xmlForm;
                    $optionsArray['instance']   = $thisContenedorAyuda->getInstance();

                    //$postOptions = Html::javascriptObject($optionsArray);
                    $jsonOptions = htmlspecialchars(json_encode($optionsArray));

                    $popayudaImg = 'popAyuda(\'DIV'.$xmlOrig.'\' , $(this).prev() , \''.$xmlHLP.'\',  event  , '.$jsonOptions.');';


//die($jsonOptions);
                    $inputBox->addParameter('helpData', $jsonOptions );
                    //$inputBox->addEvent('onkeypress', $popayuda);
              //      $inputBox->addEvent('onkeydown', $popayuda);
                    $inputBox->sufijo =  '<img  src="../img/contexthelp.png" id="img'.$id.'" onClick="'.$popayudaImg.'">';
                }

                if ($esClaveBusqueda && $modoAbm != 'readonly' &&  
                  ( isset($this->deshabilitado) && $this->deshabilitado != 'true' || !isset($this->deshabilitado))
                 &&
                        $uiClass->Datos->tipoAbm == 'ficha') {

                    $searchFunction = ' buscar( \''.$idContenedor.'\', \''.$fieldName.'\', this.value, \'=\' , \''.$uiClass->Datos->xml.'\'  , \''.$uiClass->Datos->xmlOrig.'\' , this);';
                    $inputBox->addEvent('onchange'  , $searchFunction, true);
                    //$inputBox->addEvent('onblur'  , $searchFunction, true);
                }

                if ($uiClass->Datos->tipoAbm=='cabecera') {
                    // Chequear que no exista el llenocabecera ya en el Evento (Duplicacion de Evento)
                    //$inputBox->addEvent('onblur', 'llenoCabecera(\''.$formName.'\', \''.$uiClass->Datos->xml.'\' , \''.$uiClass->Datos->xmlOrig.'\', this.value);', true);
                    if ($this->TipoDato == 'check') {
                        $inputBox->addEvent('onchange', 'llenoCabecera(\''.$formName.'\', \''.$uiClass->Datos->xml.'\' , \''.$uiClass->Datos->xmlOrig.'\', this);', true);
                    }
                }

                // Si el campo activa otros campos
                if (isset($this->activa) && $this->activa =='true' ||
                    isset($this->desactiva) && $this->desactiva =='true'
                ) {
                    $aMostrar= $uiClass->Datos->camposaMostrar();
                    $aActivar='';
                    if ($aMostrar !='')
                        foreach ($aMostrar as $campoaMostrar) {

                            if (isset($uiClass->Datos->getCampo($campoaMostrar)->activador) && trim($uiClass->Datos->getCampo($campoaMostrar)->activador) == trim($fieldName))
                                $aActivar[]=$campoaMostrar;
                        }
                    if ($aActivar !='') {
                    if ($this->desactiva=='true') 
                	 $desactiva = ', true';
                    else $desactiva = ', false';
                        $inputBox->addEvent('onchange',"Histrix.activateCheck( this ,  new Array('" . implode("','", $aActivar) . "'))".$desactiva, true);
                        $inputBox->addParameter('class', 'activate');
                        $javascripAejecutar .= '$("#'.$this->uid2.'").change();';
                    }
                }

                if ($deshabilitado == 'true') {
                    $inputBox->removeEvent(array('onfocus', 'onblur', 'onkeyup', 'onclick'));
                }


                $inputBox->sufijo .=  $icoFile.$browse.$linkExt;

                $textBoxString = $inputBox->show();


                $formatoCampo = (isset($this->Formato))?$this->Formato:null;

                // internal link in form must be non Editable
       	        if (isset($this->linkint) && $this->linkint != '' && $this->editable == 'false') {
                     $textBoxString = $uiClass->linkButton($this, $valor, $valor, $param, $formatoCampo);
        	    }

                $salida .= $browsePref.$textBoxString;
            break;
        }

        /////////////////////////////////////////////////////////////////////
        // Asigno el tabindex al objeto para cuando se refresca no perderlo
        // 
        $uiClass->Datos->getCampo($this->NombreCampo)->lastTabindex = $tabindex;

        //////////////////////////////////////////////////////////
        // add Javascript code to execute on render Field
        // 
        if (trim($javascripAejecutar) != '') {
            $salida .= Html::scriptTag(array($javascripAejecutar));
        }

        /////////////////////
        // add Label
        // 
        if (isset($this->conLabel) && $this->conLabel == 'true' && $uiClass->Datos->editable !='true') {
            $salida =  $this->renderLabel($uid2, false).' '.$salida;
        }
        return $salida;
    }

    /**
     * build field Label
     
     * @param  string $id uniqid for field
     * @return string
     */
    public function renderLabel($uid2, $rendercell=true){


        $label = new Html_label($this->Etiqueta, $uid2);


        $label->addParameter('id', $this->idlbl)->addParameter('for', $uid2);


        $styleStr = '';


        // STYLE
        if (isset($this->noshow) && $this->noshow == 'true') {
            $label->addStyle('display', 'none');
            $tdProperties['style'] = 'display:none;';
        } else {
            if (isset($this->style) && $this->style != '')
                $styleStr = $this->style;

            if (isset($this->Formstyle))
                $styleStr .= $this->Formstyle;
            $tdProperties['style'] = $styleStr;

            $label->addParameter('style', $styleStr, true);
        }

        if (isset($this->xmletiq)) {
            $contetiq = 'histrixLoader.php?xml=' . $this->xmletiq;
            $padre = ($this->_DataContainerRef->xmlOrig != '') ? 'DIV' . $this->_DataContainerRef->xmlOrig : 'DIV' . $this->_DataContainerRef->xml;
            $jsetiq = 'Histrix.loadInnerXML(\'' . $this->xmletiq . '\', \'' . $contetiq . '\', \'\',\'' . ucfirst($this->Etiqueta) . '\', \'' . $padre . '\',\'' . $this->NombreCampo . '\' );';
            $label->addEvent('onclick', $jsetiq);
            $label->addParameter('class', 'boton');
        }

        if ((isset($this->sincelda) && $this->sincelda == 'true') || !$rendercell) {
            $output = $label->show();

        } else {
            if (isset($this->rowspan))
                $tdProperties['rowspan'] = $this->rowspan;

            $tdProperties['class'] = 'lblclass';
            $output = Html::tag('td', $label->show(), $tdProperties);
        }

        return $output;

    }
}
