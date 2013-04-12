/*
 * Copyright (C) 2007 The Android Open Source Project
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package com.example.android.nfcsecurescan;

import java.io.BufferedReader;
import java.io.IOException;
import java.io.InputStream;
import java.io.InputStreamReader;
import java.io.UnsupportedEncodingException;
import java.net.MalformedURLException;
import java.net.URL;
import java.net.URLConnection;
import java.text.SimpleDateFormat;
import java.util.ArrayList;
import java.util.Date;
import java.util.List;
import java.util.Locale;

import org.apache.http.HttpEntity;
import org.apache.http.HttpResponse;
import org.apache.http.NameValuePair;
import org.apache.http.client.ClientProtocolException;
import org.apache.http.client.HttpClient;
import org.apache.http.client.entity.UrlEncodedFormEntity;
import org.apache.http.client.methods.HttpGet;
import org.apache.http.client.methods.HttpPost;
import org.apache.http.impl.client.DefaultHttpClient;
import org.apache.http.message.BasicNameValuePair;
import org.apache.http.params.BasicHttpParams;
import org.apache.http.params.HttpParams;

import com.example.android.nfcsecurescanapp.R;

import android.annotation.SuppressLint;
import android.app.Activity;
import android.content.Intent;
import android.nfc.FormatException;
import android.nfc.NdefMessage;
import android.nfc.NdefRecord;
import android.nfc.NfcAdapter;
import android.nfc.Tag;
import android.nfc.tech.Ndef;
import android.nfc.tech.NdefFormatable;
import android.os.Bundle;
import android.os.Parcelable;
import android.view.Menu;
import android.view.MenuItem;
import android.view.View;
import android.view.View.OnClickListener;
import android.widget.TextView;

/**
 * This class provides a basic demonstration of how to write an Android
 * activity. Inside of its window, it places a single view: an EditText that
 * displays and edits some internal text.
 */
public class NFCSecureScanActivity extends Activity {
    
    static final private int BACK_ID = Menu.FIRST;
    static final private int CLEAR_ID = Menu.FIRST + 1;
    static final private int DISABLE_ID = Menu.FIRST + 2;
    
    private boolean scanEnabled=true;
    private boolean setToWrite=false;
    private String messageScanned ="";
    private String serverReply="";
    private TextView tv1;
    
    public NFCSecureScanActivity() {
    }

    /** Called with the activity is first created. */
    @Override
    public void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        // Inflate our UI from its XML layout description.
        setContentView(R.layout.skeleton_activity);
        tv1 = (TextView) findViewById(R.id.textView1);
    }

    /**
     * Called when the activity is about to start interacting with the user.
     */
    @Override
    protected void onResume() {
        super.onResume();
        Intent intent = getIntent();
        NdefMessage msgs[] = null;
       
        if (NfcAdapter.ACTION_NDEF_DISCOVERED.equals(getIntent().getAction()) && scanEnabled) { 	
        	if (!setToWrite) {
        		readTagAndCallWebserver(intent,msgs);     
        		setToWrite=true;
        	} else {
        		try {
					writeServerReplyToTag();
				} catch (IOException e) {
					// TODO Auto-generated catch block
					e.printStackTrace();
				} catch (FormatException e) {
					// TODO Auto-generated catch block
					e.printStackTrace();
				}
        		setToWrite=false;
        	}
        	//now that we have submitted the tag, we want to write the response
            //that we received from the server to the tag that we scanned
            //we prompt the user to scan the tag again
            
            tv1.setText("We have received a response from the webserver, " +
            		"so in order to complete the protocol we need to scan " +
            		"the tag again so that the next person can use it. Note: " +
            		"if you do not complete this step your scan will be invalid.\nWhen " +
            		"you are ready, please go to the menu, and press the write button. " +
            		"Bring your phone up to the tag so that the data can be written.");	            
        }
    }

	private void writeServerReplyToTag() throws IOException, FormatException {
		tv1.setText("\nWe are ready to write, " +
				"please scan the tag now.");
		Intent intent = getIntent();
		Tag tag = intent.getParcelableExtra(NfcAdapter.EXTRA_TAG);
		Ndef ndef = Ndef.get(tag);
		NdefRecord record[] = new NdefRecord[1];
		record[0] = createRecord();
		NdefMessage message = new NdefMessage(record);
		if (ndef != null) {
		  ndef.connect();
		  ndef.writeNdefMessage(message);
		} else {
		  NdefFormatable format = NdefFormatable.get(tag);
		  if (format != null) {
		    format.connect();
		    format.format(message);
		  }           
		}
	}
	
	private NdefRecord createRecord() throws UnsupportedEncodingException {
	    String text       = "Hello, World!";
	    String lang       = "en";
	    byte[] textBytes  = text.getBytes();
	    byte[] langBytes  = lang.getBytes("US-ASCII");
	    int    langLength = langBytes.length;
	    int    textLength = textBytes.length;
	    byte[] payload    = new byte[1 + langLength + textLength];

	    // set status byte (see NDEF spec for actual bits)
	    payload[0] = (byte) langLength;

	    // copy langbytes and textbytes into payload
	    System.arraycopy(langBytes, 0, payload, 1,              langLength);
	    System.arraycopy(textBytes, 0, payload, 1 + langLength, textLength);

	    NdefRecord record = new NdefRecord(NdefRecord.TNF_WELL_KNOWN, 
	                                       NdefRecord.RTD_TEXT, 
	                                       new byte[0], 
	                                       payload);

	    return record;
	}
    
    private void readTagAndCallWebserver(Intent intent,NdefMessage[] msgs) {
    	Date now = new Date();
    	Parcelable[] rawMsgs = intent.getParcelableArrayExtra(NfcAdapter.EXTRA_NDEF_MESSAGES);
        if (rawMsgs != null) {
            msgs = new NdefMessage[rawMsgs.length];
            for (int i = 0; i < rawMsgs.length; i++) {
                msgs[i] = (NdefMessage) rawMsgs[i];
            }
        }
        
        NdefRecord record = msgs[0].getRecords()[0];

        byte[] payload = record.getPayload();
        messageScanned = new String(payload);
        
        //strip "en" from beginning	
        messageScanned = messageScanned.substring(3, messageScanned.length()); 
        
        tv1.setText("We scanned a tag at time :"+now.toString()+ 
        			" data was: " + messageScanned +"\n");
        
        //Now that we have scanned the message, we want to send it
        //to the web server.
        try {
			submitMessageToWeb();
		} catch (IOException e) {
			e.printStackTrace();
			tv1.setText("Failed to submit tag to webserver. It might be down.");
		}

    }

	private void submitMessageToWeb() throws IOException {
    	HttpClient httpclient = new DefaultHttpClient();
        HttpPost httppost = new HttpPost("http://nfcsecurity.appspot.com/submit");

        List<NameValuePair> nameValuePairs = new ArrayList<NameValuePair>(2);
        nameValuePairs.add(new BasicNameValuePair("password", messageScanned));
        httppost.setEntity(new UrlEncodedFormEntity(nameValuePairs));

        // Execute HTTP Post Request
        HttpResponse response = httpclient.execute(httppost);

        tv1.setText(tv1.getText()+"response : "+response.toString());
        serverReply = response.toString();
	}

	/**
     * Called when your activity's options menu needs to be created.
     */
    @Override
    public boolean onCreateOptionsMenu(Menu menu) {
        super.onCreateOptionsMenu(menu);

        // We are going to create two menus. Note that we assign them
        // unique integer IDs, labels from our string resources, and
        // given them shortcuts.
        menu.add(0, BACK_ID, 0, R.string.back).setShortcut('0', 'b');
        menu.add(0, CLEAR_ID, 0, "Write to Tag").setShortcut('1', 'c');
        menu.add(0, DISABLE_ID, 0, R.string.disable).setShortcut('2','s');
        return true;
    }

    /**
     * Called right before your activity's option menu is displayed.
     */
    @Override
    public boolean onPrepareOptionsMenu(Menu menu) {
        super.onPrepareOptionsMenu(menu);

        // Before showing the menu, we need to decide whether the clear
        // item is enabled depending on whether there is text to clear.
        menu.findItem(CLEAR_ID).setVisible(tv1.getText().length() > 0);
        return true;
    }

    /**
     * Called when a menu item is selected.
     */
    @Override
    public boolean onOptionsItemSelected(MenuItem item) {
        switch (item.getItemId()) {
        case BACK_ID:
            finish();
            return true;
        case CLEAR_ID:
            try {
				writeServerReplyToTag();
			} catch (IOException e) {
				// TODO Auto-generated catch block
				e.printStackTrace();
			} catch (FormatException e) {
				// TODO Auto-generated catch block
				e.printStackTrace();
			}
            return true;
        case DISABLE_ID:
        	disableScan();
        	return true;
        }

        return super.onOptionsItemSelected(item);
    }
    
    OnClickListener mDisableListener = new OnClickListener() {
		public void onClick(View v) {
			disableScan();
		}
    };
    
    public void disableScan() {
    	tv1.setText("Tag scan is disabled");
		scanEnabled=false;
    }
    
    
}

