


function checkDate(obj)
{
    var valor = obj.value;
	if (isDate(valor) == false)
	    {
			obj.select();
			obj.focus();
		    return false;
	}
	else
	{
	    return true;
        }
}


function checkVoid(obj)
{
    var valor = obj.value;
	if (isVoid(valor) == false)
	    {
			obj.select();
			obj.focus();
		    return false;
	}
	else
	{
	    return true;
        }
}


function isVoid(val){
	if (val==''){
		alert("El Campo no puede ser Vacio");
		return false;
	}
return true;
}

function checkNumeric(obj,minval, maxval,comma,period,hyphen)
{
    var numberfield = obj;


	if (chkNumeric(obj,minval,maxval,comma,period,hyphen) == false)
	    {

			obj.select();
			obj.focus();
		    return false;
	}
	else
	{
	    return true;
        }
}
							    
function chkNumeric(obj,minval,maxval,comma,period,hyphen){
	// only allow 0-9 be entered, plus any values passed
	// (can be in any order, and don't have to be comma, period, or hyphen)
	// if all numbers allow commas, periods, hyphens or whatever,
	// just hard code it here and take out the passed parameters
	var checkOK = "0123456789" + comma + period + hyphen;
	var checkStr = obj;
	var allValid = true;
	var decPoints = 0;
	var allNum = "";
						    
	for (i = 0;  i < checkStr.value.length;  i++){
	    ch = checkStr.value.charAt(i);
	    for (j = 0;  j < checkOK.length;  j++)
		    if (ch == checkOK.charAt(j))
		    break;
			if (j == checkOK.length){
			    allValid = false;
			    break;
		    }
		    if (ch != ",")
		    allNum += ch;
	}
	if (!allValid){	
	
		alertsay = "Los valores validos son \""
		alertsay = alertsay + checkOK + "\" en el campo \"" + obj.id + "\"."
		alert(alertsay);
		return (false);
	}
							    
	// set the minimum and maximum
	var chkVal = allNum;
	var prsVal = parseInt(allNum);
	if (chkVal != "" && !(prsVal >= minval && prsVal <= maxval)){
		alertsay = "Por valor ingrese un valor igual "
		alertsay = alertsay + "o mayor que \"" + minval + "\" o igual "
		alertsay = alertsay + "o menos que \"" + maxval + "\" en en campo \"" + checkStr.name;
		alert(alertsay);
		return (false);
	}
}							    

/**
 * DHTML date validation script. Courtesy of SmartWebby.com (http://www.smartwebby.com/dhtml/)
 */
// Declaring valid date character, minimum year and maximum year
var dtCh= "/";
var minYear=1900;
var maxYear=2100;

function isInteger(s){
	var i;
    for (i = 0; i < s.length; i++){   
        // Check that current character is number.
        var c = s.charAt(i);
        if (((c < "0") || (c > "9"))) return false;
    }
    // All characters are numbers.
    return true;
}

function stripCharsInBag(s, bag){
	var i;
    var returnString = "";
    // Search through string's characters one by one.
    // If character is not in bag, append to returnString.
    for (i = 0; i < s.length; i++){   
        var c = s.charAt(i);
        if (bag.indexOf(c) == -1) returnString += c;
    }
    return returnString;
}

function daysInFebruary (year){
	// February has 29 days in any year evenly divisible by four,
    // EXCEPT for centurial years which are not also divisible by 400.
    return (((year % 4 == 0) && ( (!(year % 100 == 0)) || (year % 400 == 0))) ? 29 : 28 );
}
function DaysArray(n) {
	for (var i = 1; i <= n; i++) {
		this[i] = 31
		if (i==4 || i==6 || i==9 || i==11) {this[i] = 30}
		if (i==2) {this[i] = 29}
   } 
   return this
}

function isDate(dtStr){
	var daysInMonth = DaysArray(12)
	var pos1=dtStr.indexOf(dtCh)
	var pos2=dtStr.indexOf(dtCh,pos1+1)
	var strDay=dtStr.substring(0,pos1)
	var strMonth=dtStr.substring(pos1+1,pos2)
	var strYear=dtStr.substring(pos2+1)
	strYr=strYear
	if (strDay.charAt(0)=="0" && strDay.length>1) strDay=strDay.substring(1)
	if (strMonth.charAt(0)=="0" && strMonth.length>1) strMonth=strMonth.substring(1)
	for (var i = 1; i <= 3; i++) {
		if (strYr.charAt(0)=="0" && strYr.length>1) strYr=strYr.substring(1)
	}
	month=parseInt(strMonth)
	day=parseInt(strDay)
	year=parseInt(strYr)
	if (pos1==-1 || pos2==-1){
		alert("El formato debe ser : dd/mm/aaaa")
		return false
	}
	if (strMonth.length<1 || month<1 || month>12){
		alert("Ingrese un mes valido")
		return false
	}
	if (strDay.length<1 || day<1 || day>31 || (month==2 && day>daysInFebruary(year)) || day > daysInMonth[month]){
		alert("Ingrese un dia valido")
		return false
	}
	if (strYear.length != 4 || year==0 || year<minYear || year>maxYear){
		alert("Ingrese un año valido de 4 digitos entre "+minYear+" y "+maxYear)
		return false
	}
	if (dtStr.indexOf(dtCh,pos2+1)!=-1 || isInteger(stripCharsInBag(dtStr, dtCh))==false){
		alert("Fecha Invalida")
		return false
	}
return true
}


							    
							    