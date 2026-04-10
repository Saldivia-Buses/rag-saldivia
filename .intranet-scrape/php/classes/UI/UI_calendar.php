<?php

/*
 * 2011-10-25
 * help popup class
 */
include "../lib/RRule/RRule.php";
//include "../lib/When/When.php";

class UI_calendar extends UI_crud {

    /**
     * User Interfase constructor
     *
     */
    public function __construct(&$Datacontainer) {
        parent::__construct($Datacontainer);
        $this->tag = 'div';
        $this->hasForm = true;
        $this->defaultClass = 'consulta';
        $this->sidePanel = (isset($this->Datos->sidePanel) && $this->Datos->sidePanel != '') ? $this->Datos->sidePanel : true;
    }

    private function fullCalendarInit($id, $events='') {

        $aspectRatio = ($this->Datos->calendarAspectRatio) ? 'aspectRatio:' . $this->Datos->calendarAspectRatio . ',' : '';


        $jscode .= "var calendar$id= $('#$id').addClass('fullcalendar').fullCalendar({
			editable: true,
                        selectable:true,
                        selectHelper:true,
                        $aspectRatio
        		header: {
				left: 'prev,next today',
				center: 'title',
				right: 'month,agendaWeek,agendaDay'
			},";
//        if ($this->Datos->modifica != 'false') {
        $jscode .= "eventClick: function(calEvent, jsEvent, view) {
                    var dataObject = calEvent.dataObject;        ";
        
    // external detail
    if (isset($this->Datos->detalle)) {
        $linkdir = (isset($this->Datos->dirxml) && $this->Datos->dirxml != '')?'&dir='.$this->Datos->dirxml:'';
        //$linkdir = (isset($ObjCampo->linkdir) && $ObjCampo->linkdir != '')?'&dir='.$ObjCampo->linkdir : $linkdir;

        if (isset($this->Datos->detailDir)){
            $linkdir =  '&dir='.$this->Datos->detailDir;
        }

	    $det  = "xml={$this->Datos->detalle}$linkdir";
	
        $fieldlist = $this->Datos->camposaMostrar();

        $this->getDataFields($fieldlist);
        foreach ($fieldlist as $order => $fieldName) {
            $Field = $this->Datos->getCampo($fieldName);
            $det .= '&'.$Field->getUrlVariableString('\'+ dataObject.'.$fieldName.'+\'', false);
	}                                   
	
	$jscode .= "    Histrix.loadInnerXML( '{$this->Datos->detalle}',
                                'histrixLoader.php?$det',
                                '',
                                'Evento',
                                '{$this->Datos->idxml}',
                                'Evento',
                                {modal:true}
                                );
";    
	
    }        
  
    $jscode .= "  if (calEvent.editable == false) return;
                    dataObject.instance = '{$this->Datos->getInstance()}';
                    $('#DIVFORM{$this->Datos->idxml}').slideDown().css({'z-Index':'100'});

                    dataObject.{$this->fieldStartDate} = $.fullCalendar.formatDate( calEvent.start, 'dd/MM/yyyy'); 
                    ";
        if ($this->fieldEndDate != '') {
            $jscode .= "dataObject.{$this->fieldEndDate}   = $.fullCalendar.formatDate( calEvent.end  , 'dd/MM/yyyy');";
        }
        $jscode .= "llenoForm(null, 'Form$id', '{$this->Datos->xml}', true, null, dataObject);
                    
                    },";

        $updateFunction .= "function(calEvent, jsEvent, view) {
                             var dataObject = calEvent.dataObject;
                             $.post('setData.php?modificar=true&xmldatos={$this->Datos->xml}&instance={$this->Datos->getInstance()}' ,dataObject, function(){

                                dataObject.{$this->fieldStartDate} = $.fullCalendar.formatDate( calEvent.start, 'yyyy-MM-dd'); ";
        if ($this->fieldStartTime != '')
            $updateFunction .= " dataObject.{$this->fieldStartTime} = $.fullCalendar.formatDate( calEvent.start, 'HH:mm:ss');";
        if ($this->fieldStartTime != '')
            $updateFunction .= " dataObject.{$this->fieldEndTime}   = $.fullCalendar.formatDate( calEvent.end  , 'HH:mm:ss');";
        if ($this->fieldStartTime != '')
            $updateFunction .= " dataObject.{$this->fieldEndDate}   = $.fullCalendar.formatDate( calEvent.end  , 'yyyy-MM-dd');";


        $updateFunction .= "$.post('process.php?accion=update&xmldatos={$this->Datos->xml}&instance={$this->Datos->getInstance()}', dataObject);
                                });
                $('#eventList$id').empty();
                        },";

        $jscode .= "eventDrop:$updateFunction";
        $jscode .= "eventResize:$updateFunction";
        $jscode .= "viewDisplay: function(view) {
                $('#eventList$id').empty();
                $('#$id').fullCalendar( 'rerenderEvents' );},";
        $jscode .= "windowResize: function(view) {
                $('#eventList$id').empty();
                $('#$id').fullCalendar( 'rerenderEvents' );},";


        $jscode .= "eventRender:function( event, element, view ) {
				    if (typeof tinyTips == \"undefined\"){
					    $.requireScript('../javascript/jquery.tinyTips.js', function() {
					        element.tinyTips('light', event.fulltitle);
					});
				    }
				    else {
				        element.tinyTips('light', event.fulltitle);
				    }

                var eventLi = jQuery('<li class=\"fc-event fc-event-skin fc-event-hori fc-corner-left fc-corner-right \">'+ $.fullCalendar.formatDate( event.start, 'dd/MM') +' - ' + event.fulltitle+'</li>')
                            .css({'background-color':event.color})
                            .bind('click',  function() {
                                if (event.editable == false) return;
                                var dataObject = event.dataObject;
                                $('#DIVFORM{$this->Datos->idxml}').slideDown().css({'z-Index':'100'});
                                llenoForm(null, 'Form$id', '{$this->Datos->xml}', true, null, dataObject);

                                });;
                    $('#eventList$id').append(eventLi);
                },";

        if ($this->Datos->inserta != 'false')
            $jscode .= "select: function(start, end, allDay) {
                            var htxdate = $.fullCalendar.formatDate( start, 'dd/MM/yyyy' );
                            $('[name={$this->fieldStartDate}]', '#DIVFORM{$this->Datos->idxml}' ).val(htxdate);
                            $('#DIVFORM{$this->Datos->idxml}').slideDown().css({'z-Index':'100'}); },";

        $jscode .= "monthNames: ['Enero', 'Febrero', 'Marzo', 'Abril', 'Mayo', 'Junio', 'Julio', 'Augosto', 'Septiembre', 'Octubre', 'Noviembre', 'Diciembre'],
                        monthNamesShort: ['Ene', 'Feb', 'Mar', 'Abr', 'May', 'Jun', 'Jul', 'Ago', 'Sep', 'Oct', 'Nov', 'Dic'],
                        dayNames: ['Domingo', 'Lunes', 'Martes', 'Miercoles', 'Jueves', 'Viernes', 'Sabado'],
                        dayNamesShort: ['Dom', 'Lun', 'Mar', 'Mie', 'Jue', 'Vie', 'Sab'],
                        buttonText:{
                            today:    'Hoy',
                            month:    'mes',
                            week:     'semana',
                            day:      'día'
                        },
                        dragOpacity: {
                            ''   : .7
                        },
                       columnFormat: {
                            month: 'ddd',    
                            week: 'ddd d/M', 
                            day: 'dddd d/M'  
                        },
                        titileFormat: {
                            month: 'MMMM yyyy',                            
                            week: \"d MMM [ yyyy]{ '&#8212;'d [ MMM] yyyy}\",
                            day: 'dddd, d MMM, yyyy'
                        },
                        eventSources: [ 'process.php?reload=true&xmldatos=".$this->Datos->xml."&instance=". $this->Datos->getInstance()."' ]
                    }); ";

        //$jscode .= "calendar$id.fullCalendar( 'rerenderEvents' );";

// 		$.requireScript(['../javascript/fullcalendar/fullcalendar.js','../javascript/fullcalendar/gcal.js'], function(){

	$finaljscode = "if (typeof fullcalendar == \"undefined\"){
		$.requireScript(['../javascript/fullcalendar/fullcalendar.js'], function(){
		    $jscode
		});
	}
	else {
		$jscode	
	
	}
	
	";
        return $finaljscode;
    }

    // get data Fields from XML
    private function getDataFields($fieldlist) {
        foreach ($fieldlist as $order => $fieldName) {
            $Field = $this->Datos->getCampo($fieldName);
            $dataType = Types::getTypeXSD($Field->TipoDato);

            if ($Field->calendarId)
                $this->fieldId = $fieldName;

            if ($Field->calendarStartDate)
                $this->fieldStartDate = $fieldName;
            if ($Field->calendarEndDate)
                $this->fieldEndDate = $fieldName;
            if ($Field->calendarColor)
                $this->fieldColor = $fieldName;
            if ($Field->calendarEditable)
                $this->fieldEditable = $fieldName;
            if ($Field->calendarStartTime)
                $this->fieldStartTime = $fieldName;
            if ($Field->calendarEndTime)
                $this->fieldEndTime = $fieldName;
            if ($Field->calendarRecurringRule) {
                $this->fieldRecurringRule = $fieldName;
                $Field->addClass = 'recurringrule';
            }

            if ($Field->calendarSubject) {

                $this->fieldSubject = $fieldName;
                if (isset($Field->opcion)) {
                    $this->fieldArray = $Field->opcion;
                }
            }
            if ($Field->calendarDescription)
                $this->fieldDescription = $fieldName;

            if ($dataType == 'xsd:date' && $this->fieldStartDate == '') {
                $this->fieldStartDate = $fieldName;
            }
            if ($dataType == 'xsd:date'){
              $this->dates[$fieldName]=true;
            }

            if ($dataType == 'xsd:string' && $this->fieldSubject == '') {
                $this->fieldSubject = $fieldName;
            }
        }
    }

    public function getIcalDate($time, $incl_time = true) {
        return $incl_time ? date('Ymd\THis', $time) : date('Ymd', $time);
    }

    public function createEvents($startParam='',$endParam='') {

        $this->Datos->Select();
        $this->Datos->CargoTablaTemporal();

        $fieldlist = $this->Datos->camposaMostrar();

        $this->getDataFields($fieldlist);
        $separator = '';

        if (is_array($this->Datos->TablaTemporal->datos()))
        {
            foreach ($this->Datos->TablaTemporal->datos() as $rowNumber => $row) 
            {
                $recurringRule = $row[$this->fieldRecurringRule];

                $startDate = $row[$this->fieldStartDate];
                $startTime = $row[$this->fieldStartTime];
                $endDate   = $row[$this->fieldEndDate];
                $endTime   = $row[$this->fieldEndTime];
                $description = $row[$this->fieldDescription];
                $color 	   = $row[$this->fieldColor];
                $editable  = ($row[$this->fieldEditable]) ? true : false; // Boolean Type
                $start     = $startDate;
                $end       = $endDate;


                $allDay = true;
                if ($startTime != '' && $startTime != '00:00:00') {
                    $start = $startDate . 'T' . $startTime . 'Z';
                    $allDay = false;
                }

                if ($endDate == '0000-00-00')
                    $endDate = $startDate;

                $startTimestamp = strtotime($startDate);

                // skip future events
                if ($endParam != ''){
                    if ($startTimestamp > $endParam) continue;
                }

                if ($endTime != '') {
                    $end = $endDate . 'T' . $endTime . 'Z';
                }

                $id = $row[$this->fieldId];
                $value = $row[$this->fieldSubject];

                if (isset($this->fieldArray)) {
                    $value = $this->fieldArray[$value];
                    if (is_array($value))
                        $value = current($value);
                }

                foreach($this->dates as  $fName => $val)
                {
                    $row[$fName] = date($this->dateFormat, strtotime($row[$fName]));
                }

                // this will unhide invisible items
                if ($value == '') $value = $description.' ';

                $title = htmlspecialchars($value);
                $title = (strlen($title) > 40) ? substr($title, 0, 40) . '...' : $title;

                $row[$this->fieldStartDate] = date($this->dateFormat, strtotime($row[$this->fieldStartDate]));

                if ($this->fieldEndDate != '') 
                {
                    if ($row[$this->fieldEndDate] != '0000-00-00') 
                    {
                        $row[$this->fieldEndDate] = date($this->dateFormat, strtotime($row[$this->fieldEndDate]));
                    } 
                    else unset($row[$this->fieldEndDate]);
                }

                $row[$this->fieldStartDate] = date('Y-m-d', strtotime($start));

                if ($title != '')
                {

                    if ($recurringRule != '') 
                    {
                        $loopcounter = 0;
                        $RuleParser = new RRule($this->getIcalDate(strtotime($start)), $recurringRule);

                        while ($result = $RuleParser->getNext()) 
                        {
                            $loopcounter++;
                            $year = $result->render('Y')  ;

                            if (date('Y') + 1 < $year && $loopcounter > 10) break;

                            $start = $result->render('Y-m-d');

                            $hash = date('Y-m-d', strtotime($start));
                            $row[$this->fieldStartDate] = date('Y-m-d', strtotime($start));

                            $event_object = [
                                "id" => $id,
                                "title" => $title,
                                "fulltitle" => $description,
                                "allDay" => $allDay,
                                "start" => $start,
                                "end" => $end,
                                "editable" => $editable,
                                "dataObject" => $row,
                                "color" => $color
                            ];

                            $myarray[] = $event_object;
                        }
                    } 
                    else 
                    {
                            $event_object = [
                                "id" => $id,
                                "title" => $title,
                                "fulltitle" => $description,
                                "allDay" => $allDay,
                                "start" => $start,
                                "end" => $end,
                                "editable" => $editable,
                                "dataObject" => $row,
                                "color" => $color ];

                            $myarray[] = $event_object;
                    }

                }
            }
        }

        return json_encode($myarray, JSON_HEX_QUOT);
    }

    public function getGoogleCalendar($url='') {


        if ($url == '') {
            $db = $_SESSION["db"];
            $database = Cache::getCache('datosbase' . $db);
            $url = $database->calendar;
        }
        $event = "{ url:'$url' ,
                className: 'gcal-event',
                target:'_blank',
                textColor: 'white',
                color: 'green'}";
        return $event;
    }

    // render de complete XML
    public function show($idFormulario = '', $divcont='', $opt='') {

        $id = 'Show' . $this->Datos->idxml;

        // id del contenedor (creo)
        $id2 = str_replace('.', '_', ($divcont != '') ? $divcont : $id);

        $style = (isset($this->Datos->style)) ? $this->Datos->style : '';

        // Columns
        if ($this->sidePanel === true)
            $ancho = (isset($this->Datos->ancho)) ? $this->Datos->ancho : '70';
        else
            $ancho = 100;

        $width = (isset($this->Datos->width)) ? $this->Datos->width : $ancho;

        if ($width != null) {
            $this->Datos->col1 = $width;
            $this->Datos->col2 = 100 - $width;
            $style.='width:' . $this->Datos->col1 . '%;';
        }
        $clasedetalle = 'detalle2';
        $clase = $this->defaultClass;
        $salidaDatos = $this->showTabla();

//      if ($sidePanel)
//          $barraSlide = $this->showSlider($id);

        $clase_der = (isset($this->formClass)) ? 'class="' . $this->formClass . '"' : 'class="consulta_der"';

        $salidaAbm = $this->showAbm(null, $clase_der);


        // Si se define explicitamente una clase en el xml
        if ($this->Datos->clase != '') {
            $clase = $this->Datos->clase;
        }

        $retorno = '';
        if ($this->Datos->campoRetorno != '') {
            $uidRetorno = $this->Datos->getCampo($this->Datos->campoRetorno)->uid;
            $retorno = ' origen="' . $uidRetorno . '" ';
        }

        $salida  = '<div class="' . $clase . '" id="' . $id . '" style="' . $style . '" ' . $retorno . '>';
        $salida .= '<div class="contewin" >';
        $salida .= $salidaDatos;
        $salida .= '</div>';
        $salida .= '</div>';

        // El Abm
        $salida .= $salidaAbm;

        // Incorporo la barra vertical para slide
        $salida .= $barraSlide;

        if ($this->sidePanel === true)
            $salida .= $this->sidePanel();

        // Add Detail div
        //$salida .= $this->detailDiv($clasedetalle);

        // create Javascript functions
        $script[] = "Histrix.registroEventos('" . $this->Datos->idxml . "')";
//      $script[] = "Histrix.calculoAlturas('" . $this->Datos->idxml . "', null ); ";
        $salida .= Html::scriptTag($script);

        return $salida;
    }

    public function showTablaInt($opt = '', $idTabla = '', $segundaVez = '', $nocant='', $div=false, $form=null, $pdf=null, &$parentObject=null) {
  /*      $llenoTemporal = (isset($this->Datos->llenoTemporal)) ? $this->Datos->llenoTemporal : '';
        $preload = (isset($this->Datos->preloadData)) ? $this->Datos->preloadData : '';

        if ($llenoTemporal != "false" && $segundaVez == '' && $opt != 'noselect') {
            if ($this->nosel != 'true') {
                if (!isset($this->Datos->tempTables[$llenoTemporal])) {

                    if ($preload != "false") {
                        $this->Datos->Select();
                    }
                    $this->Datos->preloadData = "true";

                    if ($this->Datos->resultSet)
                        $this->cantCampos = _num_fields($this->Datos->resultSet);
                }
                $this->Datos->CargoTablaTemporal();
            }
        }
*/
        $this->createEvents();


        $jscode[] = $this->fullCalendarInit($this->Datos->idxml, $events);
        if ($this->Datos->calendarTasks != '')
            $jscode[] = "$(\"#tasks{$this->Datos->idxml}\").load(\"histrixLoader.php?xml={$this->Datos->calendarTasks}&dir={$this->Datos->dirxml}&xmlsub=test\", {'__inline':true, '__inlineid':'{$this->Datos->idxml}'});";


        $output = Html::scriptTag($jscode);

        echo $output;
    }

    private function sidePanel() {
        $title = date('D, d M Y');
        $style = '';
        $strStyle = '';
        if ($this->Datos->col2 != '')
            $style .='width:' . ($this->Datos->col2 - 0.5) . '%;';

        if ($style != '')
            $strStyle = ' style="' . $style . '" ';


        $output .= "<div $strStyle class=\"consulta_der\">";
        //    $output .= "<h1>$title</h1>";
        $output .= "<div class=\"panel\">";
        $output .= "<ul id=\"eventList{$this->Datos->idxml}\" class=\"eventList\" ><li>Procesando..</li></ul>";
        $output .= '</div>';
        if ($this->Datos->calendarTasks != '') {
            $output .= "<div id=\"tasks{$this->Datos->idxml}\" class=\"panel\">";
            $output .= '</div>';
        }
        $output .= '</div>';
        return $output;
    }

}

?>