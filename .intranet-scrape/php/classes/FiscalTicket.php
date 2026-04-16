<?php
/*
 * Creado el 2007/07/17
 *
 * Implementacion dela Clase para impresion en controlador fiscal
 */
class FiscalTicket extends Fiscal {


    var $stIn;
    var $stOut;
    var $archivo;
    var $renglones;

    function constructor($stIn='', $stOut='') {

    }

    // Cierre Z o X
    function cierre($tipo) {
        $this->DailyClose($tipo);
        $this->Imprimir();
    }

    function sanitize($nombre) {
        $nombre = filter_var($nombre, FILTER_SANITIZE_STRING, FILTER_FLAG_STRIP_HIGH);
        $nombre = str_replace('á','a',$nombre);
        $nombre = str_replace('é','e',$nombre);
        $nombre = str_replace('í','i',$nombre);
        $nombre = str_replace('ó','o',$nombre);
        $nombre = str_replace('ú','u',$nombre);
        $nombre = str_replace('ñ','n',$nombre);
        $nombre = str_replace('Ñ','ñ',$nombre);

        $nombre = str_replace('Á','A',$nombre);
        $nombre = str_replace('É','E',$nombre);
        $nombre = str_replace('Í','I',$nombre);
        $nombre = str_replace('Ó','O',$nombre);
        $nombre = str_replace('Ú','U',$nombre);

        return $nombre;
    }

    function OpenFiscalReceipt($tipoC=''){
        $this->renglones[]=array('OpenTicket' => $tipoC);
    }

    function addRenglon($renglon) {
        $this->renglones[] = $renglon;
    }
    function SetHeaderTrailer($linea, $texto) {
        $this->renglones[]=']'.chr(0x1c).$linea.chr(0x1c)
            .$texto.chr(0x1c)
            ."\n";

    }
    function printLineItem($desc, $cantidad, $unitario, $iva) {
        $desc= substr($this->sanitize($desc),0 , 20);
        $cantidad = sprintf("%08s", number_format($cantidad , 2) * 1000);
        
        $unitario = sprintf("%08s", number_format($unitario , 2) * 100);
        $iva      = sprintf("%04s", number_format($iva , 2) * 100);
         /*
        * Descripción del artículo.
        * Cantidad.
        * Precio unitario.
        * Tasa de IVA.
        * Calificador de item: M,m,R,r
        * Cantidad de bultos
        * Impuestos internos porcentuales
        * Impuestos Internos Fijos (opcional)
        */

        $string = $desc.'|'.$cantidad.'|'.$unitario.'|'.$iva .'|M'.'|00001|00000000|000000000';
        
        $this->renglones[]=array('SendTicketItem' => $string );
    }

/*
        // ABRO CONTROLADOR PARA LEER RESPUESTA
        $fdIn = fopen($this->stIn,'r+b');
        if (!$fdIn) {
            throw new Exception('Error al abrir Controlador Fiscal para Lectura');
        }
        $this->fdIn= $fdIn;


        // Abro el dispositivo y escribo en la impresora
        $fdOpen = fopen($this->stOut,'w+b');
        $errorCode = '';
        foreach ($this->renglones as $nlinea => $linea) {
            fwrite($fdOpen, $linea);

            // ESCRIBO UN RENGLON Y LEO LA RESPUESTA
            $respuesta = $this->readAnswer($linea, 50);
            // Error stop execution
            if ($this->isError($respuesta)) {
                $errorCode=$this->lastError;
            }

        }
        fclose($fdOpen);

        fclose($fdIn); // Cierro el archivo de respuesta


*/
    function getConfigurationData() {
        $this->renglones[]='f'.chr(0x1c)."\n";
    }


    function Subtotal() {
    //    $this->renglones[]=array('GetTicketSubtotal' => 'P| ' );
    }

    /**
    * SendTicketPayment: Envía un pago en un ticket fiscal
    * Texto descriptivo del pago.
    * Monto a pagar.
    * Descripción del pago: T=Pago; t=Reversión del pago; D=Descuento; R=Recargo.
    **/
    function sendTiquetPayment($desc , $importePago,  $type) {

        $importePago = sprintf("%08s", number_format(abs($importePago) , 2) * 100);
        $string = $desc.'|'.$importePago.'|'.$type;
        $this->renglones[]=array('SendTicketPayment' => $string );

    }




    function addPago($importePago, $desc) {

        $this->renglones[]='D.'.$desc.'.'.$importePago.'.T.0';

    //D.Efectivo.100.00.T.0
    }

    function cierreFactura() {
        $this->renglones[]='E'."\n";
    }

    function GeneralDiscount($discount='') {
        if ($discount == '') return;

        if ($discount != 0) {
            $texto = 'BONIFICACION';
            $imputacion = 'm';
        }
/*		if ($discount > 0){
			$texto = 'RECARGO';
			$imputacion = 'M';
		}*/
        if ($imputacion != '')
            $this->renglones[]='T'.chr(0x1c).$texto.chr(0x1c).$discount.chr(0x1c).$imputacion .
                chr(0x1c).'0'.chr(0x1c).'T'."\n";
    }

    function CloseFiscalReceipt() {
        $this->renglones[]=array('CloseTicket' => '' );
    }

    function DailyClose($tipo) {
            $this->renglones[]=array('CloseJournal' => $tipo );
    }

    function CancelTicket() {
            $this->renglones[]=array('CancelTicket' => '' );
            $this->Imprimir();
    }

    function Imprimir() {
    // Creo un Archivo temporal primero
        $fname = tempnam("/tmp", 'compfiscal.fis');
        $fh=fopen($fname, "r+b");

        foreach ($this->renglones as $nlinea => $linea) {
            foreach($linea as $command => $options){
                $opt = ($options != '')?' --options:"'.$options.'"':'';
                $string = '/usr/local/bin/rbfiscal --serialport:1 --command:'.$command.$opt."\n";
                try{
                    loger($string, 'rbfiscal.log');
                    $result = shell_exec($string);
                    loger($result, 'exec.log');
                }
                catch(Exception $ex){
                    loger('error executing: '.$string, 'exec_error.log');
                }
            }
            fwrite($fh, $string);
        }
        fclose($fh);
        $this->filename = $fname;

        // HAGO UNA COPIA DE SEGURIDAD DEL COMPROBANTE
        $uid= uniqid('fiscal');
        @copy($fname, '../tmp/'.$uid.'.fis');
        //unlink($fname);

        if ($errorCode != '') return false;

        return true;

    }


    function isError($answer) {

        $ok = true;

        // busco por Cadena de Error.
        if ( strstr($answer, 'c080') !== false &&
            strstr($answer, '0600') !== false) {
            $ok = true;
        }

        if ( strstr($answer, 'c080') !== false &&
            strstr($answer, '3600') !== false) {
            $ok = true;
        }

        if ( strstr($answer, 'c080') !== false &&
            strstr($answer, '8610') !== false) {
            $ok = true;
        }


        if ( strstr($answer, 'c080') !== false &&
            strstr($answer, '8620') !== false) {
            $this->lastError = 'ERROR EN LOS DATOS ENVIADOS AL CONTROLADOR FISCAL';
            $ok = false;
        }

        if ($ok === false) {
            $this->logFiscal('ERROR');

            return true;
        }
        return false;
    }

    function logFiscal($string='') {
        $fecha = date('d/m/Y H:i:s');
        $datapath    = $_SESSION["datapath"];

        if ($datapath == '')
            $datapath    = $_SESSION["db"]; // TODO remove tomorrow

        $dir = '../database/'.$datapath.'/log/';
        $archivo = $dir.'Fiscal.log';
        $recurso = fopen($archivo, 'a+');

        fwrite($recurso, $fecha.' '.$string);
        fclose($recurso);
    }

    function readAnswer($origin, $length=50) {
    //		$respuesta =fread($this->fdIn, $length);
        $respuesta =fgets($this->fdIn);

        $string = '['.$origin.']: '.$respuesta;
        $this->logFiscal($string);

        return $respuesta;

    }

    function getStatus() {
        $comando="*\n"; // Comando para obtener Status

        try {
            $fdOut = fopen($this->stOut,'w+b');
            if (!$fdOut) {
                throw new Exception('Error al abrir Controlador Fiscal');
            }
            $this->fdOut= $fdOut;
            $respuesta = fwrite($this->fdOut, $comando);
            if ($respuesta === FALSE )
                throw new Exception('Error al enviar comando al Controlador Fiscal');

            fclose($fdOut);

            $fdIn = fopen($this->stIn,'r+b');
            if (!$fdIn) {
                throw new Exception('Error al abrir Controlador Fiscal para Lectura');
            }
            $this->fdIn= $fdIn;
            //echo 'in'.$this->fdIn;
            $respuesta = $this->readAnswer('Status',60);
            //$respuesta =fread($this->fdIn, 50);

            if ($respuesta === FALSE) {
                throw new Exception('Error al abrir Controlador Fiscal para Lectura');
            }
            fclose($fdIn);

        // leo hasta que encuentro el codigo de finalizacion de linea

/*
			do {
				$char = fgetc($this->fdIn);

			    $respuesta .= $char;
			} while (ord($char) != 10 );
*/
        //die( '<div class="error" >Buscando Respuesta:'. $respuesta.'</div>');



        } catch(Exception $e) {
            echo '<div class="error" > Exception: ',  $e->getMessage(), "</div>";
            die ();

        }

        $arrayCadenas = split(chr(0x1c),$respuesta);
        $this->statusString = $arrayCadenas;
        //print_r($arrayCadenas);
        //die('buscando');
        return $respuesta;
    }

    // Falta implementar NC
    function getNumerador($letra) {
        $array = $this->statusString;
        switch($letra) {
            case 'A':
                return $array[4];
                break;
            case 'B':
                return $array[2];
                break;
        }
    }
}

?>
