<?php
set_time_limit(300);
$DirectAccess = true;

include ("./autoload.php");
include_once('./inicio_sesion_remota.php');
include ("./sessionCheck.php");
include("../lib/feedCreator/feedcreator.class.php");

$titulo=$_GET['titulo'];

//$xmldatos=$_GET['xmldatos'];
//$xmlOrig=$_REQUEST['xmlOrig'];
//$subdir = $_GET["dir"];
$xmldatos = 'calendar.xml';
$subdir	  = 'histrix/calendar/';
$dirXML = '../xml/';

$xmlPath=$datosbase->xmlPath;
$dirXML = '../database/'.$xmlPath.'xml/';

$titulo = $xmlPath;
//$MisDatos = leoXML($dirXML, $xmldatos, null, null, $subdir);

$xmlReader = new Histrix_XmlReader($dirXML, $xmldatos, null, null, $subdir);

if ($ContenedorReferente != '')
    $xmlReader->addReferentContainer($ContenedorReferente);

$xmlReader->addParameters($_GET);
$xmlReader->addParameters($_POST);

$MisDatos   = $xmlReader->getContainer();

$UI = 'UI_'.str_replace('-', '', 'consulta');
$datos = new $UI($MisDatos);

$datos->setTitulo($MisDatos->tituloAbm);
$noprint= $datos->show();

$TZID = 'America/Argentina/Buenos_Aires';
// ICAL HEADER
$ics_contents  = "BEGIN:VCALENDAR\n";
$ics_contents .= "VERSION:2.0\n";
$ics_contents .= "PRODID:HISTRIX CALENDAR\n";
$ics_contents .= "METHOD:PUBLISH\n";
$ics_contents .= "X-WR-CALNAME:{$MisDatos->tituloAbm}\n";
 
 # Change the timezone as well daylight settings if need be
 
 $ics_contents .= "X-WR-TIMEZONE:$TZID\n";
 $ics_contents .= "BEGIN:VTIMEZONE\n";
 $ics_contents .= "TZID:$TZID\n";
 $ics_contents .= "X-LIC-LOCATION:$TZID\n";
 $ics_contents .= "BEGIN:STANDAR\n";
 $ics_contents .= "TZOFFSETFROM:-0300\n";
 $ics_contents .= "TZOFFSETTO:-0300\n";
 $ics_contents .= "TZNAME:ART\n";
 $ics_contents .= "DTSTART:19700101T000000\n";
// $ics_contents .= "RRULE:FREQ=YEARLY;BYMONTH=3;BYDAY=2SU\n";
 //$ics_contents .= "TZNAME:EDT\n";
 $ics_contents .= "END:STANDAR\n";
 $ics_contents .= "END:VTIMEZONE\n";
                                 /*
 $ics_contents .= "BEGIN:STANDARD\n";
 $ics_contents .= "TZOFFSETFROM:-0400\n";
 $ics_contents .= "TZOFFSETTO:-0500\n";
 $ics_contents .= "DTSTART:20071104T020000\n";
 $ics_contents .= "RRULE:FREQ=YEARLY;BYMONTH=11;BYDAY=1SU\n";
 $ics_contents .= "TZNAME:EST\n";
 $ics_contents .= "END:STANDARD\n";
 $ics_contents .= "END:VTIMEZONE\n";
                                   */
                                                 /*
$MisDatos->Select();
$MisDatos->CargoTablaTemporal();
                                               */
$Tablatemp = $MisDatos->TablaTemporal->datos();

 foreach($Tablatemp as $row => $schedule_details){

   $id            = md5($schedule_details['id_calendar']);
   $start_time    = $schedule_details['starttime'];
   $end_date      = $schedule_details['EndDate'];
   $end_time      = $schedule_details['endtime'];
   $category      = $schedule_details['Category'];
   $name          = $schedule_details['Name'];
   $location      = $schedule_details['Location'];
   $description   = $schedule_details['subject'];
                    
   // Remove '-' in $start_date and $end_date
   $estart_date   = str_replace("-", "", $start_date);
   $eend_date     = str_replace("-", "", $end_date);
                           
   // Remove ':' in $start_time and $end_time
   $estart_time   = str_replace(":", "", $start_time);
   $eend_time     = str_replace(":", "", $end_time);
                                  
   // Replace some HTML tags
   $search  = array("<br>","&amp;","&rarr;" , "&larr;", ",", ";");
   $replace = array("\\n" ,"&","-->" , "<--", "\\,", "\\;");

   $location      = str_replace($search, $replace,   $location);
   $description   = str_replace($search, $replace,   $description);

    /*                                                 
   $location      = str_replace("<br>", "\\n",   $location);
   $location      = str_replace("&amp;", "&",    $location);
   $location      = str_replace("&rarr;", "-->", $location);
   $location      = str_replace("&larr;", "<--", $location);
   $location      = str_replace(",", "\\,",      $location);
   $location      = str_replace(";", "\\;",      $location);
                                                              
   $description   = str_replace("<br>", "\\n",   $description);
   $description   = str_replace("&amp;", "&",    $description);
   $description   = str_replace("&rarr;", "-->", $description);
   $description   = str_replace("&larr;", "<--", $description);
   $description   = str_replace("<em>", "",      $description);
   $description   = str_replace("</em>", "",     $description);
   */                                                                        
     // Change TZID if need be
   $ics_contents .= "BEGIN:VEVENT\n";
   $ics_contents .= "DTSTART;VALUE=DATE:".date('Ymd',strtotime($start_time))."\n";
   $ics_contents .= "DTEND;VALUE=DATE:".date('Ymd',strtotime($start_time))."\n";
   //TZID=New_York"     . $estart_date . "T". $estart_time . "\n";
//   $ics_contents .= "DTEND:"       . $eend_date . "T". $eend_time . "\n";
//   $ics_contents .= "DTSTAMP:"     . date('Ymd') . "T". date('His') . "Z\n";
   $ics_contents .= "LOCATION:"    . $location . "\n";
   $ics_contents .= "DESCRIPTION:" . $description . "\n";
   $ics_contents .= "SUMMARY:"     . $description . "\n";
   $ics_contents .= "UID:"         . $id . "\n";
   $ics_contents .= "SEQUENCE:0\n";
   $ics_contents .= "STATUS:CONFIRMED\n";

   $ics_contents .= "END:VEVENT\n";
 }

$ics_contents  .= "END:VCALENDAR\n";

header("Content-Disposition:attachment; filename=\"".$titulo.".ics\"");

echo $ics_contents;
 
/*
$mixml = new Cont2XML($MisDatos, $xmldatos, null, null, true, true, null);
$mixml->generateFeed();

header ("Content-type: text/xml");
//header("Content-Disposition:attachment; filename=\"".$titulo.".xml\"");
echo $mixml->show();
*/
?>