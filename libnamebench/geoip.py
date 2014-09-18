# Copyright 2010 Google Inc. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

"""Class used for determining GeoIP location."""

import re
import tempfile

# external dependencies (from nb_third_party)
import httplib2
import simplejson

import util


def GetFromFreegeoip():
    apiUrl = 'http://www.freegeoip.net/json'
    h = httplib2.Http()
    response, content = h.request(apiUrl, 'GET')
    results = json.loads(content)
    output = ''
    order = {
        'country': 1,
        'city': 2,
        'region_name': 3,
        'region_code': 4,
        'longitude': 5,
        'latitude': 6,
        'ip': 7,
    }
    return results

def GetGeoData():
  """Get geodata from any means necessary. Sanitize as necessary."""
  try:
    json_data = GetFromFreegeoip()
    # if not json_data:
      # json_data = GetFromGoogleLocAPI()

    # Make our data less accurate. We don't need any more than that.
    json_data['latitude'] = '%.3f' % float(json_data['latitude'])
    json_data['longitude'] = '%.3f' % float(json_data['longitude'])
    return json_data
  except:
    print 'Failed to get Geodata: %s' % util.GetLastExceptionString()
    return {}
