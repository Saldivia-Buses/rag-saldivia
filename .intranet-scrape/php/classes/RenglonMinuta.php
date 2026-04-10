<?php
 /**
  * clase para cada renglon de la minuta contable
  * */

 class RenglonMinuta {

    var $TBLminuta = 'CTBREGIS';

    var $doh;           // 1 - debe / 2 - Haber
    var $cuenta;        // codigo de Cuenta Contable
    var $importe;       // Importe
    var $referencia;    // Referencia del renglon
    var $valido;
    var $habil;         // Cuenta habilitada o no
    var $tipo;          // Indica si es o no automatico
    var $ctbcentro_id;  // centro de costos
    var $nombre_centro; 

    public function RenglonMinuta($cuenta, $doh, $importe, $nombre='', $tipo='', $ctbcentro_id= ''){
        $this->cuenta    = $cuenta;
        $this->doh       = $doh;
        $this->importe   = $importe;
        $this->nombre    = $nombre;
        $this->tipo      = $tipo;
        $this->ctbcentro_id      = $ctbcentro_id;
        $this->valido    = true;
        $this->habil     = true;

        if ($ctbcentro_id != '') {
            $sql = 'select nombre_centro from CTB_CENTROS where id_ctbcentro = '.$ctbcentro_id;
            $rs = consulta($sql, null);
            while ($row = _fetch_array($rs)) {
                $this->nombre_centro = $row['nombre_centro'];
            }
        }

    }

 }

?>
