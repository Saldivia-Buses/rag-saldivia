<?php
class Histrix_Table {
    var $nombreTabla;

    /* Array de campos que almacena*/
    var $campos;
    var $indices;
  //  var $camposTabla; // lista de campos total de la tabla
   // var $ArrayCampos; // Array con las propiedades de los campos de la tabla

    public function __construct($NombreTabla, $buscoindex = 'true') {
        $this->nombreTabla = $NombreTabla;

        if ($NombreTabla != '') {
        // get Cache Table Metadata
            $key = $_SESSION['db'].'.'.$NombreTabla;
            $ArrayCampos = Cache::getCache($key);

            if ($ArrayCampos === false) {

                if (function_exists('campos'))
                    $camposTabla = campos($NombreTabla);

                //Lleno el Array de Campos del SQL (Necesario para obtener los tipos)
                if (function_exists('_fetch_array')){
                    while ($fila = _fetch_array($camposTabla)) {
                        $ArrayCampos[$fila['Field']]= $fila;
                    }
                }
                // set Cache Table Metadata
                Cache::setCache($key, $ArrayCampos);
            }


            $this->cacheFields = $ArrayCampos ;

            //	busco los indices de la tabla
            if ($buscoindex == 'true') {

                if ($ArrayCampos != '')
                    foreach($ArrayCampos as $nomcampo => $fila) {
                        if ($fila['Key']=='PRI') {
                            $campos[$nomcampo] = $fila['Field'];
                        }
                    }
                if (isset($campos))
                    $this->indices = $campos;
            }
        //$this->indices = getCamposIndice($NombreTabla);
        }
    }

    public function getFieldMetadata(){
        $key = $_SESSION['db'].'.'.$this->nombreTabla;
        $fieldMetaData =  Cache::getCache($key);
        return $fieldMetaData;
    }
    public function getNombre() {
        return $this->nombreTabla;
    }

    public function addOpcion($campo, $valor, $Desc) {
        $this->campos[$campo]->addOpcion($valor, $Desc);
    }

    /*
     * Agrego un campo a la tabla actual
     */
    public function addCampo($campo, $Exp = '', $Etiq = '', $formato = '', $TablaPadre = '', $local = '') {

        if ($TablaPadre == '') $TablaPadre = $this->nombreTabla;
        if ($Etiq 		== '') $Etiq = $campo;

        $this->campos[$campo] = new Field($campo, $Exp, $Etiq, $formato, $TablaPadre);
        if ($local != '')
        $this->campos[$campo]->local = $local;

        $this->etiquetas_reverse[$Etiq] = $campo;

        if ($Exp == '' && $local == '' && isset($this->indices[$campo])) {
            $this->campos[$campo]->esClave = true;
        }

        $ArrayCampos = $this->cacheFields;

        if (isset($ArrayCampos[$campo])) {

            $nomTipo = $ArrayCampos[$campo]["Type"];

            $ini= strpos ($nomTipo, '(');
            if ($ini !== false) {
                $end= strpos ($nomTipo, ')');
                $size = substr($nomTipo, $ini +1 , $end - 1 - $ini  );
                $this->campos[$campo]->Size = $size;

                $nomTipo =  substr($nomTipo, 0 , $ini  );
            }
            $this->campos[$campo]->TipoDato = $nomTipo;

            $xsdType = Types::getTypeXSD($nomTipo,'-');

            // Los campos fechas les agrego 2 espacios para los años de 4 cifras
            if ($xsdType == 'xsd:date')
                $this->campos[$campo]->Size += 2;
        }

    }

    /*
     * Metodo que obtiene un campo a partir de su nombre, si no lo encuentra
     * usa la etiqueta para identificarlo
     */
    public function getCampo($nombre) {
        if ($nombre =='') return null;

        if (isset($this->campos[$nombre])) {
            return $this->campos[$nombre];
        }
        else {
            if (isset($this->etiquetas_reverse[$nombre]) && $this->etiquetas_reverse[$nombre]  != '') {
                if ($this->campos[$this->etiquetas_reverse[$nombre]] != null)
                    return $this->campos[$this->etiquetas_reverse[$nombre]];
            }

            if ($this->campos)
                foreach ($this->campos as $n => $Dest) {
                    if ($nombre == $Dest->Etiqueta) {
                        return $Dest;
                    }
                }
        }

        /*
         * Si no tengo Campos agrego en la lista de campos el que tenga el mismo nombre en el
         * Select de la tabla (o algo asi)
         */
        if ($this->cantCampos() < 1) {
            if (isset($ArrayCampos) && is_array($ArrayCampos))
            foreach ($ArrayCampos as $Nro => $Items) {

                $fieldName=(isset($Items["COLUMN_NAME"]))?$Items["COLUMN_NAME"]:$Items["Field"];
                $this->addCampo($fieldName);
                return $this->campos[$fieldName];
            }
        }
        return null;
    }

    public function &getCampoRef($nombre) {
        if ($nombre =='') return null;

        /*
         * Si no tengo Campos agrego en la lista de campos el que tenga el mismo nombre en el
         * Select de la tabla (o algo asi)
         */

        /*if (($this->campos[$nombre] != null)) {
                return $this->campos[$nombre];
        }*/

        if (array_key_exists($nombre, $this->campos))
            return $this->campos[$nombre];
        else {
            if ($this->campos)
                foreach ($this->campos as $n => $Dest) {
                    if ($nombre == $Dest->Etiqueta) {
                        return $Dest;
                    }
                }
        }
        if ($this->cantCampos() < 1) {
            foreach ($ArrayCampos as $Nro => $Items) {
            //$fieldName=(isset($Items["COLUMN_NAME"]))?$Items["COLUMN_NAME"]:$Items["Field"];
                $fieldName=$Items["Field"];
                $this->addCampo($fieldName);
                return $this->campos[$fieldName];
            }
        }
        $null = null;
        return $null;
    }

    /*
     * Seteo la etiqueta de un Campo
     */
    public function setEtiqueta($nombre, $Etiqueta) {
        $this->getCampo($nombre)->Etiqueta = $Etiqueta;
        $this->etiquetas_reverse[$Etiqueta] = $nombre;

    }

    /*
     * Devuelvo la cantidad de campos que tiene la tabla para usar en los ABM's'
     */
    public function cantCampos() {
        return count($this->campos);
    }

}
?>