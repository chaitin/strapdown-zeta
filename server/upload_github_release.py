#!/usr/bin/env python
try:
    from urllib.parse import quote
except ImportError:
    from urllib import quote
import mimetypes
import os
import os.path
import requests

class GitHubUnexpectedResponse( RuntimeError ):
    def __init__( self, resp ):
        self.resp = resp

class Repo:
    def __init__( self, owner, repo, token=None, base='https://api.github.com' ):
        self.base = base
        self.owner = owner
        self.repo = repo
        self.authorization = 'token '+token if token else None

    def request( self, method, uri, base=None, *args, **kwargs ):
        headers = kwargs.setdefault( 'headers', {} )
        headers.setdefault( 'Accept', 'application/vnd.github.v3+json' )
        if self.authorization:
            headers.setdefault( 'Authorization', self.authorization )
        kwargs.setdefault( 'allow_redirects', False )
        if base is None:
            base = self.base
        url = '%s/repos/%s/%s%s' % ( base, self.owner, self.repo, uri )
        return requests.request( method, url, *args, **kwargs )

    def _get( self, *args, **kwargs ):
        return self.request( 'GET', *args, **kwargs )

    def _post( self, *args, **kwargs ):
        return self.request( 'POST', *args, **kwargs )

    def _patch( self, *args, **kwargs ):
        return self.request( 'PATCH', *args, **kwargs )

    def _delete( self, *args, **kwargs ):
        return self.request( 'DELETE', *args, **kwargs )

    def get_list( self, *args, **kwargs ):
        resp = self._get( *args, **kwargs )
        if resp.status_code == 404:
            return []
        elif resp.status_code == 200:
            return resp.json()
        else:
            raise GitHubUnexpectedResponse( resp )

    def get_single( self, *args, **kwargs ):
        resp = self._get( *args, **kwargs )
        if resp.status_code == 404:
            return None
        elif resp.status_code == 200:
            return resp.json()
        else:
            raise GitHubUnexpectedResponse( resp )

    def create( self, *args, **kwargs ):
        resp = self._post( *args, **kwargs )
        if resp.status_code == 201:
            return resp.json()
        else:
            raise GitHubUnexpectedResponse( resp )

    # https://developer.github.com/v3/repos/releases/

    def list_release( self, page=1 ):
        '''
        Information about published releases are available to everyone. Only users
        with push access will receive listings for draft releases.
        '''
        query = {}
        if page > 1:
            query['page'] = page
        return self.get_list( '/releases', params=query )

    def get_release_by_id( self, release_id ):
        return self.get_single( '/releases/%d' % release_id )

    def get_latest_release( self ):
        '''
        View the latest published full release for the repository. Draft releases
        and prereleases are not returned by this endpoint.
        '''
        return self.get_single( '/releases/latest' )

    def get_release_by_tag( self, tag_name ):
        '''
        Get a published release with the specified tag.
        '''
        return self.get_single( '/releases/tags/' + quote(tag_name) )

    '''
    release attributes

    Name                Type        Description
    tag_name            string      Required. The name of the tag.
    target_commitish    string      Specifies the commitish value that determines where the Git tag is created from. Can be any branch or commit SHA. Unused if the Git tag already exists. Default: the repository's default branch (usually master).
    name                string      The name of the release.
    body                string      Text describing the contents of the tag.
    draft               boolean     true to create a draft (unpublished) release, false to create a published one. Default: false
    prerelease          boolean     true to identify the release as a prerelease. false to identify the release as a full release. Default: false
    '''

    def create_release( self, tag_name, target_commitish=None, name=None, body=None,
                        draft=None, prerelease=None ):
        '''
        Users with push access to the repository can create a release.
        '''
        data = { k:v for k,v in locals().items() if v is not None }
        data.pop('self')
        return self.create( '/releases', json=data )

    def edit_release( self, release_id, tag_name=None, target_commitish=None, name=None, body=None,
                      draft=None, prerelease=None ):
        '''
        Users with push access to the repository can edit a release.
        '''
        data = { k:v for k,v in locals().items() if v is not None }
        data.pop('self')
        data.pop('release_id')
        return self._patch( '/releases/%d' % release_id, json=data )

    def delete_release( self, release_id ):
        '''
        Users with push access to the repository can delete a release.
        '''
        return self._delete( '/releases/%d' % release_id )

    def list_release_asset( self, release_id ):
        return self.get_list( '/releases/%d/assets' % release_id )

    def upload_release_asset( self, release_id, data, name=None, label=None ):
        if name is None:
            if hasattr( data, 'name' ):
                name = os.path.basename( data.name )
            else:
                raise ValueError( 'upload_release_asset: name is required' )
        query = { k:v for k,v in locals().items() if v is not None }
        query.pop('self')
        query.pop('release_id')
        query.pop('data')
        content_type = mimetypes.guess_type( name )[0] or 'application/octet-stream'
        headers = { 'Content-Type': content_type }
        return self.create( '/releases/%d/assets' % release_id, data=data, params=query, headers=headers, base='https://uploads.github.com' )

    def get_release_asset( self, asset_id ):
        return self.get_single( '/releases/assets/%d' % asset_id )

    def download_release_asset( self, asset_id ):
        return self._get( '/releases/assets/%d' % asset_id,
            headers={ 'Accept': 'application/octet-stream' },
            allow_redirects=True,
            stream=True,
        )

    '''
    release asset attributes

    Name    Type    Description
    name    string  Required. The file name of the asset.
    label   string  An alternate short description of the asset. Used in place of the filename.
    '''

    def edit_release_asset( self, asset_id, name=None, label=None ):
        data = { k:v for k,v in locals().items() if v is not None }
        data.pop('self')
        data.pop('asset_id')
        return self._patch( '/releases/assets/%d' % asset_id, json=data )

    def delete_release_asset( self, asset_id ):
        return self._delete( '/releases/assets/%d' % asset_id )

def main():
    oauth_token = os.environ['GITHUB_OAUTH_TOKEN']
    os_name = os.environ['TRAVIS_OS_NAME']
    tag_name = os.environ['TRAVIS_TAG']
    commit_hash = os.environ['TRAVIS_COMMIT']
    short_commit_hash = commit_hash[:7]
    filename = '/tmp/strapdown-server_%s_%s.%s.zip' % ( os_name, tag_name, short_commit_hash )

    repo = Repo( 'chaitin', 'strapdown-zeta', token=oauth_token )

    try:
        release = repo.create_release( tag_name )
    except GitHubUnexpectedResponse:
        release = repo.get_release_by_tag( tag_name )

    os.chdir( os.path.dirname( os.path.realpath( __file__ )))
    os.spawnlp( os.P_WAIT, 'zip', '-9', filename, 'strapdown-server' )
    with open( filename ) as f:
        repo.upload_release_asset( release['id'], f, label='Binary (%s)'%os_name )

if __name__ == '__main__':
    main()
