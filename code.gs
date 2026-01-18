var spreadsheet=SpreadsheetApp.openByUrl("https://docs.google.com/spreadsheets/d/1tAUYXrqy0JmojU6z6wRiW6okPs-0WKfwXrZOiMLNAxo/edit?gid=0#gid=0");
var sheet=spreadsheet.getSheetByName("details")
var payment=spreadsheet.getSheetByName("payment")
var columns = ['id','full_name', 'phone_number','Age', 'Gender', 'chief_complaint', 'present_history', 'medical_history', 'observation', 'palpation', 'examination', 'rehab', 'diagnosis', 'created_time', 'updated_time','last_paid_amount','status'];

function doPost(e) {
  Logger.log("Request: "+e)
  var methodOverride = e.parameter['X-HTTP-Method-Override'];
  if (methodOverride === 'PATCH') {
    return handleUpdateRequest(e)
  } else {
    if(methodOverride === 'POST') {
      var task=e.parameter['task']
      if(task === 'PATIENT_DETAILS' || task == null)
       return createData(e)
      else if(task === 'PAYMENT_DETAILS') {
        type= e.parameter['type']
        if(type === 'CREATE') {
          const response = savePaymentDetails(JSON.parse(e.postData.contents));
          return ContentService.createTextOutput(JSON.stringify(response))
                 .setMimeType(ContentService.MimeType.JSON);
        }else if(type === 'UPDATE') {
          const response = updatePaymentDetails(JSON.parse(e.postData.contents));
                return ContentService.createTextOutput(JSON.stringify(response))
        .setMimeType(ContentService.MimeType.JSON);
        }else if(type === 'DELETE') {
          var paymentUniqueId=e.parameter['payment_ref_id']
          return deletePaymentByUniqueId(paymentUniqueId)
        }
      } else{
        return "INVALID"
      }
    }
   
  }
}

//-------------------------------PATCH REQUEST----------------------------------------------

function handleUpdateRequest(e) {
  var jsonData = JSON.parse(e.postData.contents);
    var id = jsonData.id;
    Logger.log("updating id: "+id)
    var updated = updateData(sheet, id, jsonData);
    if (updated) {
      return ContentService.createTextOutput(JSON.stringify({
  message: 'Row Updated successfully'
})).setMimeType(ContentService.MimeType.JSON);

    } else {
      return ContentService.createTextOutput(JSON.stringify({
  message: 'Row not found'
})).setMimeType(ContentService.MimeType.JSON);

    }
}
function updateData(sheet, id, data) {
  Logger.log("updating id: "+id)
  var range = sheet.getDataRange();
  var values = range.getValues();
  for (var row = 0; row < values.length; row++) {
    if (values[row][0] === id) {
      var newRow = [];
      columns.forEach(function(column) {
        newRow.push(data[column] || '');
      });
      sheet.getRange(row + 1, 1, 1, newRow.length).setValues([newRow]);
      return true;
    }
  }
  return false;
}

//-------------------------------GET REQUEST----------------------------------------------

function doGet(e) {
  var task=e.parameter['task']
  if(task === 'PATIENT_DETAILS') {
    var isList=e.parameter['type'];
    if(isList === 'LIST') {
      return getList(e);
    }
    var data=getAllData(e);
    Logger.log(data)
    return data;
  }else if(task === 'PAYMENT_DETAILS') {
    var id=e.parameter['patient_id'];
    var paymentDetails;
    if(id === 'ALL') {
      paymentDetails=getAllPayment();
    }else {
      paymentDetails = getPaymentDetailsByPatientRef(id);
    }

    // Return the result as JSON
    return ContentService.createTextOutput(
      JSON.stringify({ success: true, data: paymentDetails })
    ).setMimeType(ContentService.MimeType.JSON);
  }
}
function getPaymentDetailsByPatientRef(patientRef) {
  // Read all data from the sheet
  const data = payment.getDataRange().getValues();
  const headers = data[0]; // The first row is treated as headers
  const result = [];

  // Iterate through the data to find rows matching the patient_ref
  for (let i = 1; i < data.length; i++) {
    if (data[i][0] === patientRef) { // Match the first column (patient_ref)
      const record = {};
      for (let j = 0; j < headers.length; j++) {
        record[headers[j]] = data[i][j]; // Map headers to row data
      }
      result.push(record);
    }
  }

  return result;
}

function getAllPayment() {
  // Read all data from the sheet
  const data = payment.getDataRange().getValues();

  // Check if the sheet has at least headers and one row of data
  if (data.length < 2) {
    return []; // Return an empty array if no data
  }

  const headers = data[0]; // The first row is treated as headers

  // Map the remaining rows to objects
  return data.slice(1).map(row => {
    const record = {};
    headers.forEach((header, index) => {
      record[header] = row[index] || ''; // Default to empty string for missing values
    });
    return record;
  });
}

function getAllData(e) {
  
  var data = sheet.getDataRange().getValues();
  
  // Create an array to hold the JSON objects
  var jsonData = [];
  Logger.log(data)
  // Loop through each row of data
  data.forEach(function(row, index) {
    if (index === 0) return; // Skip the header row
    
    // Create a JSON object for each row
    var jsonObject = {};
    columns.forEach(function(column, columnIndex) {
      jsonObject[column] = row[columnIndex];
    });
    
    // Add the JSON object to the array
    jsonData.push(jsonObject);
  });
   
  // Convert the array of JSON objects to a JSON string
  var jsonOutput = JSON.stringify(jsonData);
  
  Logger.log(jsonOutput)
  // Return the JSON string
  return ContentService.createTextOutput(jsonOutput).setMimeType(ContentService.MimeType.JSON);


}

function getList(e) {
  
  var data = sheet.getDataRange().getValues();
  
  var result = [];
  
  // Loop through the data, start at index 1 to skip headers
  for (var i = 1; i < data.length; i++) {
    var row = {
      id: data[i][0],           // ID is in column 1
      full_name: data[i][1],    // Full name is in column 2
      diagnosis: data[i][12]    // Diagnosis is in column 12
    };
    result.push(row);
  }
  
  // Return the data as JSON
  return ContentService.createTextOutput(JSON.stringify(result))
           .setMimeType(ContentService.MimeType.JSON);


}

//-------------------------------POST REQUEST----------------------------------------------

function createData(e) {
  var jsonData = JSON.parse(e.postData.contents);
  
  // Validate the JSON data
  if (!jsonData || typeof jsonData !== 'object') {
    return ContentService.createTextOutput(JSON.stringify({
  message: 'Invalid Json Data'
})).setMimeType(ContentService.MimeType.JSON);

  }
  
  // Get the values from the JSON data
  var values = [];
  columns.forEach(function(column) {
    values.push(jsonData[column] || '');
  });
  
  // Add the values to the sheet
  sheet.appendRow(values);
  
  // Return a success message
  return ContentService.createTextOutput(JSON.stringify({
  message: 'Data added successfully'
})).setMimeType(ContentService.MimeType.JSON);

}
//-----------------------------------SAVE PAYMENT------------------------
function savePaymentDetails(paymentData) {
  try {

    // Check if the sheet exists, create headers if it's new
    if (!payment.getLastRow()) {
      payment.appendRow(["patient_ref", "unique_payment_id", "amount", "mode", "date"]);
    }

    // Extract data from the provided object
    const { patient_ref, unique_payment_id, amount, mode, date } = paymentData;

    // Append the data to the sheet
    payment.appendRow([patient_ref, unique_payment_id, amount, mode, date]);
    

    const patientDetails = sheet.getDataRange().getValues(); // Get all data from the summary sheet
    // Loop through the rows of the summary sheet to find the patient reference and update the amount
    for (let i = 1; i < patientDetails.length; i++) { // Starting from 1 to skip header row
      if (patientDetails[i][0] === patient_ref) { // Assuming patient_ref is in the first column
        sheet.getRange(i + 1, 16).setValue(amount); // Assuming last_payment_amount is in the second column
        break;
      }
    }

    // Return success response
    return { success: true, message: "Payment details saved successfully!" };

  } catch (error) {
    // Return error response
    return { success: false, message: error.message };
  }
}
function updatePaymentDetails(paymentData) {
  try {

    // Check if the sheet exists, create headers if it's new
    if (!payment.getLastRow()) {
      payment.appendRow(["patient_ref", "unique_payment_id", "amount", "mode", "date"]);
    }

    // Extract data from the provided object
    const { patient_ref, unique_payment_id, amount, mode, date } = paymentData;
    var data = payment.getDataRange().getValues();
    var headerRow = data[0];
    
    // Find the index of the relevant columns
    var paymentIdColumn = headerRow.indexOf("unique_payment_id");
    var amountColumn = headerRow.indexOf("amount");
    var modeColumn = headerRow.indexOf("mode");
    var dateColumn = headerRow.indexOf("date");
    var isUpdated=false;
    for (var i = 1; i < data.length; i++) {
      var row = data[i];
      if (row[paymentIdColumn] === unique_payment_id) {
        // Update the payment details in the found row
        payment.getRange(i + 1, amountColumn + 1).setValue(amount);
        payment.getRange(i + 1, modeColumn + 1).setValue(mode);
        payment.getRange(i + 1, dateColumn + 1).setValue(date);
        isUpdated=true;
        break;
      }
    }

    if(isUpdated)
    // Return success response
      return { success: true, message: "Payment details updated successfully!" };
    else
       return { success: false, message: "Payment reference not found" };

  } catch (error) {
    // Return error response
    return { success: false, message: error.message };
  }
}

//------------------DELETE payment-----------------------
function deletePaymentByUniqueId(uniquePaymentId) {
  // Open the Google Sheet by its ID or use the active spreadsheet
  
  // Get all the data from the sheet
  var data = payment.getDataRange().getValues();
  
  // Loop through the rows (start from 1 to skip header row)
  // Loop through the rows (start from 1 to skip header row)
  for (var i = 1; i < data.length; i++) {
    // The unique_payment_id is assumed to be in the second column (index 1)
    if (data[i][1] === uniquePaymentId) {
      // Delete the row if the unique_payment_id matches
      payment.deleteRow(i + 1); // +1 because sheet rows are 1-indexed
      Logger.log('Deleted row ' + (i + 1)); // Optional: log for confirmation
      
      // Return a JSON response indicating success
      return ContentService.createTextOutput(
        JSON.stringify({
          success: true,
          message: 'Payment deleted successfully.',
          deleted_row: i + 1
        })
      ).setMimeType(ContentService.MimeType.JSON);
    }
  }
  
  // If no match is found, return a JSON response indicating failure
  return ContentService.createTextOutput(
    JSON.stringify({
      success: false,
      message: 'No matching unique_payment_id found.'
    })
  ).setMimeType(ContentService.MimeType.JSON);
}

