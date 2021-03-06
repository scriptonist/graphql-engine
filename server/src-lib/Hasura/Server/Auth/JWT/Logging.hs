module Hasura.Server.Auth.JWT.Logging
  ( JwkRefreshLog (..)
  , JwkFetchError (..)
  )
  where

import           Hasura.Prelude

import qualified Data.ByteString.Lazy as BL
import qualified Network.HTTP.Types   as HTTP

import           Data.Aeson
import           Network.URI          (URI)

import           Hasura.HTTP
import           Hasura.Logging       (EngineLogType (..), Hasura, InternalLogTypes (..),
                                       LogLevel (..), ToEngineLog (..))


-- | Possible errors during fetching and parsing JWK
-- (the 'Text' type at the end is a friendly error message)
data JwkFetchError
  = JFEHttpException !HttpException !Text
  -- ^ Exception while making the HTTP request
  | JFEHttpError !URI !HTTP.Status !BL.ByteString !Text
  -- ^ Non-2xx HTTP errors from the upstream server
  | JFEJwkParseError !Text !Text
  -- ^ Error parsing the JWK response itself
  | JFEExpiryParseError !(Maybe Text) Text
  -- ^ Error parsing the expiry of the JWK
  deriving (Show)

instance ToJSON JwkFetchError where
  toJSON = \case
    JFEHttpException e _ ->
      object [ "http_exception" .= e ]

    JFEHttpError url status body _ ->
      object [ "status_code" .= HTTP.statusCode status
             , "url"         .= tshow url
             , "response"    .= bsToTxt (BL.toStrict body)
             ]

    JFEJwkParseError e msg ->
      object [ "error" .= e, "message" .= msg ]

    JFEExpiryParseError e msg ->
      object [ "error" .= e, "message" .= msg ]

data JwkRefreshLog
  = JwkRefreshLog
  { jrlLogLevel :: !LogLevel
  , jrlMessage  :: !(Maybe Text)
  , jrlError    :: !(Maybe JwkFetchError)
  } deriving (Show)

instance ToJSON JwkRefreshLog where
  toJSON (JwkRefreshLog _ msg err) =
      object [ "message" .= msg
             , "error" .= err
             ]

instance ToEngineLog JwkRefreshLog Hasura where
  toEngineLog jwkRefreshLog =
    (jrlLogLevel jwkRefreshLog, ELTInternal ILTJwkRefreshLog, toJSON jwkRefreshLog)
